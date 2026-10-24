package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/authn/password"
	"github.com/bryanster/blacklight/internal/authn/session"
	"github.com/bryanster/blacklight/internal/authz"
	"github.com/bryanster/blacklight/internal/config"
	"github.com/bryanster/blacklight/internal/httpapi/gen"
	"github.com/bryanster/blacklight/internal/store"
	"github.com/bryanster/blacklight/internal/store/identity"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

// The login endpoints, through the real chain and a real temporary DuckDB.
// Nothing here reaches past the HTTP layer to set something up that a request
// could have set up instead: what is being tested is what a browser gets.

const (
	loginPath    = BasePath + "/auth/login"
	logoutPath   = BasePath + "/auth/logout"
	mePath       = BasePath + "/auth/me"
	passwordPath = BasePath + "/auth/password"

	testEmail    = "alice@example.com"
	testPassword = "correct horse battery staple"
	testNewPass  = "a different passphrase entirely"
)

// testPasswordHash is computed once for the whole package. Argon2id at the
// configured cost takes well over a hundred milliseconds by design, and every
// test in this file needs a user to exist; the verifications that are actually
// under test are not shared and still cost what they cost.
var testPasswordHash = sync.OnceValue(func() string {
	hash, err := password.Hash(testPassword)
	if err != nil {
		panic("hashing the test password: " + err.Error())
	}
	return hash
})

// authServer is a server, the database under it, and its log.
type authServer struct {
	handler http.Handler
	db      *store.DB
	logs    *logBuffer

	// manager is a session manager built from the same configuration as the
	// server's, so that a test can derive the CSRF token a session token maps
	// to (M1-005) without reaching into the server it is testing.
	manager *session.Manager
}

// newAuthServer builds the real chain over a migrated database. adjust changes
// the configuration for the tests that are about a particular setting.
func newAuthServer(t *testing.T, adjust ...func(*config.Config)) *authServer {
	t.Helper()
	return newAuthServerDeps(t, nil, adjust...)
}

// newAuthServerDeps is newAuthServer with an optional Deps mutator (fixture
// adapters, etc.).
func newAuthServerDeps(t *testing.T, adjustDeps func(*Deps), adjust ...func(*config.Config)) *authServer {
	t.Helper()

	cfg := testConfig(t)
	for _, fn := range adjust {
		fn(&cfg)
	}
	db := storetest.Migrated(t)
	var depsFns []func(*Deps)
	// Content URL validation (M7-014) resolves hostnames; default it to a stub
	// so no test touches the network. SSRF-fence tests override it explicitly.
	depsFns = append(depsFns, func(d *Deps) {
		if d.ContentLookupIP == nil {
			d.ContentLookupIP = stubContentLookupIP
		}
	})
	if adjustDeps != nil {
		depsFns = append(depsFns, adjustDeps)
	}
	handler, logs := newTestServerWith(t, cfg, db, depsFns...)

	manager, err := session.New(identity.NewSessions(db), session.OptionsFrom(cfg))
	if err != nil {
		t.Fatalf("building the test session manager: %v", err)
	}
	return &authServer{handler: handler, db: db, logs: logs, manager: manager}
}

// seedUser creates an account directly, which is what blctl does — there is
// no endpoint that creates one until M1-016.
func (s *authServer) seedUser(t *testing.T, adjust ...func(*identity.NewUser)) identity.User {
	t.Helper()

	in := identity.NewUser{
		Email:        testEmail,
		DisplayName:  "Alice",
		PasswordHash: testPasswordHash(),
		PlatformRole: authz.PlatformRoleAdmin,
		Status:       identity.StatusActive,
	}
	for _, fn := range adjust {
		fn(&in)
	}

	user, err := identity.NewUsers(s.db).Create(t.Context(), in)
	if err != nil {
		t.Fatalf("seeding the user: %v", err)
	}
	return user
}

// post sends a JSON body, optionally carrying a session cookie.
//
// A session cookie brings its CSRF cookie and header with it, because that is
// what a signed-in browser sends: the SPA's client attaches the header to every
// state-changing request (M1-005), so a helper that left it out would be
// testing a client nobody ships. The tests that are *about* CSRF build their
// requests by hand for exactly that reason — see csrf_test.go.
func (s *authServer) post(target, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	return s.send(http.MethodPost, target, body, cookies...)
}

// send is post for the methods that are not POST — PATCH and DELETE arrived
// with user administration (M1-016) — and carries the same cookies and the same
// CSRF header for the same reason.
func (s *authServer) send(method, target, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	for _, cookie := range cookies {
		request.AddCookie(cookie)
		if cookie.Name == session.CookieName {
			s.attachCSRF(request, cookie)
		}
	}
	return do(s.handler, request)
}

// attachCSRF adds the CSRF cookie and header belonging to a session cookie.
// The value is derived, not read from a previous response, so it is right even
// for a session the server has since revoked or rotated away.
func (s *authServer) attachCSRF(request *http.Request, sessionCookie *http.Cookie) {
	token := s.manager.CSRFToken(session.Token(sessionCookie.Value))
	request.AddCookie(&http.Cookie{Name: session.CSRFCookieName, Value: token})
	request.Header.Set(CSRFHeader, token)
}

func (s *authServer) get(target string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	return do(s.handler, request)
}

// login signs in and returns the response.
func (s *authServer) login(email, plaintext string) *httptest.ResponseRecorder {
	return s.post(loginPath, fmt.Sprintf(`{"email":%q,"password":%q}`, email, plaintext))
}

// signIn signs in, insists it worked, and returns the session cookie.
func (s *authServer) signIn(t *testing.T) *http.Cookie {
	t.Helper()

	recorder := s.login(testEmail, testPassword)
	if recorder.Code != http.StatusOK {
		t.Fatalf("login = %d, want 200\nbody: %s", recorder.Code, recorder.Body)
	}
	return sessionCookie(t, recorder)
}

// sessionCookie returns the session cookie a response set, failing the test if
// it set none.
func sessionCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == session.CookieName {
			return cookie
		}
	}
	t.Fatalf("no %s cookie in the response\nheaders: %v", session.CookieName, recorder.Header())
	return nil
}

// userID returns the one seeded user's identifier.
func (s *authServer) userID(t *testing.T) string {
	t.Helper()

	users, _, err := identity.NewUsers(s.db).Page(t.Context(), identity.PageFilter{Limit: 200})
	if err != nil || len(users) == 0 {
		t.Fatalf("reading the seeded user: %v", err)
	}
	return users[0].ID
}

// sessions returns every session row the seeded user has.
func (s *authServer) sessions(t *testing.T) []identity.Session {
	t.Helper()

	found, err := identity.NewSessions(s.db).ListByUser(t.Context(), s.userID(t))
	if err != nil {
		t.Fatalf("listing sessions: %v", err)
	}
	return found
}

// execSQL reaches past the repositories to age a row, which is the only way to
// reach an expiry without waiting for one.
func (s *authServer) execSQL(t *testing.T, stmt string, args ...any) {
	t.Helper()

	err := s.db.Write(t.Context(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(context.Background(), stmt, args...)
		return err
	})
	if err != nil {
		t.Fatalf("running %q: %v", stmt, err)
	}
}

// --- Signing in --------------------------------------------------------------

func TestLoginSetsTheSessionCookieAndRecordsTheLogin(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	user := server.seedUser(t)

	before := time.Now().Add(-time.Second)
	// Signed in with the address in a different case, because the account is one
	// account however it is typed.
	recorder := server.login("ALICE@Example.com", testPassword)
	if recorder.Code != http.StatusOK {
		t.Fatalf("login = %d, want 200\nbody: %s", recorder.Code, recorder.Body)
	}

	result := decodeJSON[gen.LoginResult](t, recorder)
	switch {
	case result.Status != gen.LoginStatusAuthenticated:
		t.Errorf("status = %q, want %q", result.Status, gen.LoginStatusAuthenticated)
	case result.User == nil:
		t.Fatal("the response carries no user")
	case result.User.Id != user.ID:
		t.Errorf("user.id = %q, want %q", result.User.Id, user.ID)
	case result.User.Email != testEmail:
		t.Errorf("user.email = %q, want the address as it was typed when the account was made (%q)",
			result.User.Email, testEmail)
	case result.User.PlatformRole != gen.PlatformRoleAdmin:
		t.Errorf("user.platformRole = %q, want %q", result.User.PlatformRole, gen.PlatformRoleAdmin)
	}

	cookie := sessionCookie(t, recorder)
	if cookie.Value == "" {
		t.Error("the session cookie is empty")
	}

	reread, err := identity.NewUsers(server.db).ByID(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("re-reading the user: %v", err)
	}
	if reread.LastLoginAt.Before(before) {
		t.Errorf("last_login_at = %s, want a time after %s", reread.LastLoginAt, before)
	}
}

// TestTheSessionCookieCarriesItsProtections asserts what reaches the browser,
// which is the only place these attributes matter.
func TestTheSessionCookieCarriesItsProtections(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)
	cookie := server.signIn(t)

	switch {
	case !cookie.HttpOnly:
		t.Error("the cookie is not HttpOnly")
	case !cookie.Secure:
		t.Error("the cookie is not Secure, on a production configuration")
	case cookie.SameSite != http.SameSiteStrictMode:
		t.Errorf("SameSite = %v, want Strict", cookie.SameSite)
	case cookie.Path != "/":
		t.Errorf("Path = %q, want /", cookie.Path)
	case cookie.Domain != "":
		t.Errorf("Domain = %q, want empty", cookie.Domain)
	case cookie.Expires.IsZero():
		t.Error("the cookie has no expiry; it would outlive the session as a browser session cookie")
	}
}

func TestDevelopmentIsTheOnlyThingThatDropsSecure(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t, func(cfg *config.Config) { cfg.Env = config.EnvDevelopment })
	server.seedUser(t)

	if server.signIn(t).Secure {
		t.Error("Secure was set on a development deployment, where the local http origin cannot send it")
	}
}

// TestEveryFailedLoginIsTheSameAnswer is the enumeration defence: the response
// must not say which of these it was.
func TestEveryFailedLoginIsTheSameAnswer(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)
	server.seedUser(t, func(u *identity.NewUser) {
		u.Email = "disabled@example.com"
		u.Status = identity.StatusDisabled
	})
	server.seedUser(t, func(u *identity.NewUser) {
		u.Email = "sso-only@example.com"
		u.PasswordHash = "" // an account that signs in through a provider
	})

	attempts := map[string]*httptest.ResponseRecorder{
		"wrong password":    server.login(testEmail, "not the right password"),
		"unknown address":   server.login("nobody@example.com", testPassword),
		"disabled user":     server.login("disabled@example.com", testPassword),
		"no local password": server.login("sso-only@example.com", testPassword),
	}

	var first, firstName string
	for name, recorder := range attempts {
		if recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s = %d, want 401\nbody: %s", name, recorder.Code, recorder.Body)
			continue
		}
		body := withoutInstance(t, recorder)
		if first == "" {
			first, firstName = body, name
			continue
		}
		if body != first {
			t.Errorf("%s answers\n  %s\nand %s answers\n  %s\n— the two are distinguishable",
				name, body, firstName, first)
		}
	}

	if !strings.Contains(first, string(gen.ProblemCodeUnauthenticated)) {
		t.Errorf("the failure code is not %q: %s", gen.ProblemCodeUnauthenticated, first)
	}
	for _, leak := range []string{"disabled", "no such", "not found", "password hash"} {
		if strings.Contains(strings.ToLower(first), leak) {
			t.Errorf("the response contains %q, which says why the login failed: %s", leak, first)
		}
	}
}

// withoutInstance renders a problem document with the request ID removed, so
// that two responses can be compared for being the same answer.
func withoutInstance(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()

	var problem map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decoding %s: %v", recorder.Body, err)
	}
	delete(problem, "instance")
	encoded, err := json.Marshal(problem)
	if err != nil {
		t.Fatalf("re-encoding: %v", err)
	}
	return string(encoded)
}

// TestAnUnknownAddressCostsAHashToo is the timing half of the same defence. The
// margin is deliberately wide: what it is looking for is the difference between
// an Argon2id derivation running and not running at all, which is two orders of
// magnitude and not a few percent.
//
// Each side is measured several times and the fastest run of each compared. A
// single sample of each is a coin flip on a loaded machine — both sides inflate,
// and whichever request happened to land on a busy scheduler slice decides the
// ratio — while the fastest run is the closest either side gets to its
// uncontended cost, which is the work this asserts.
func TestAnUnknownAddressCostsAHashToo(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t, func(cfg *config.Config) {
		// Several attempts per address are needed for a fastest-of comparison;
		// the account lockout would otherwise cut the wrong-password side short
		// at five. This test is about timing, not throttling.
		cfg.Throttle.AccountFailures = 100
	})
	server.seedUser(t)

	measure := func(email string) time.Duration {
		best := time.Duration(math.MaxInt64)
		for range 5 {
			start := time.Now()
			if got := server.login(email, "not the right password").Code; got != http.StatusUnauthorized {
				t.Fatalf("login = %d, want 401", got)
			}
			if elapsed := time.Since(start); elapsed < best {
				best = elapsed
			}
		}
		return best
	}

	wrongPassword := measure(testEmail)
	unknownAddress := measure("nobody@example.com")

	if unknownAddress < wrongPassword/4 {
		t.Errorf("an unknown address was answered in %s and a wrong password in %s; "+
			"the difference is an oracle for who has an account here",
			unknownAddress, wrongPassword)
	}
}

// TestTheTokenNeverLeavesTheCookie is the acceptance criterion stated directly.
func TestTheTokenNeverLeavesTheCookie(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)

	recorder := server.login(testEmail, testPassword)
	cookie := sessionCookie(t, recorder)
	token := cookie.Value

	if strings.Contains(recorder.Body.String(), token) {
		t.Errorf("the login response body contains the session token: %s", recorder.Body)
	}
	me := server.get(mePath, cookie)
	if strings.Contains(me.Body.String(), token) {
		t.Errorf("GET /auth/me contains the session token: %s", me.Body)
	}
	if logs := server.logs.String(); strings.Contains(logs, token) {
		t.Errorf("the log contains the session token:\n%s", logs)
	}

	// Nor is the value in the cookie what is kept: the database holds a hash.
	for _, stored := range server.sessions(t) {
		if stored.TokenHash == token {
			t.Error("the token was stored verbatim")
		}
	}
}

func TestLoggingInTwiceProducesTwoIndependentSessions(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)

	first := server.signIn(t)
	second := server.signIn(t)

	if first.Value == second.Value {
		t.Fatal("the second sign-in reused the first session's token")
	}
	if got := len(server.sessions(t)); got != 2 {
		t.Errorf("%d session rows, want 2", got)
	}

	if got := server.post(logoutPath, "", first).Code; got != http.StatusNoContent {
		t.Fatalf("logout = %d, want 204", got)
	}
	if got := server.get(mePath, first).Code; got != http.StatusUnauthorized {
		t.Errorf("the logged-out session = %d, want 401", got)
	}
	if got := server.get(mePath, second).Code; got != http.StatusOK {
		t.Errorf("the other session = %d, want 200 — signing out of one place is not signing out of all of them", got)
	}
}

// --- Signing out --------------------------------------------------------------

// TestLogoutRevokesTheRowAndAReplayIsRefused. The browser dropping the cookie is
// not the mechanism; the mechanism is revoked_at.
func TestLogoutRevokesTheRowAndAReplayIsRefused(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)
	cookie := server.signIn(t)

	recorder := server.post(logoutPath, "", cookie)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("logout = %d, want 204\nbody: %s", recorder.Code, recorder.Body)
	}

	cleared := sessionCookie(t, recorder)
	if cleared.Value != "" || cleared.MaxAge >= 0 {
		t.Errorf("the response did not clear the cookie: value %q, max-age %d",
			cleared.Value, cleared.MaxAge)
	}

	rows := server.sessions(t)
	if len(rows) != 1 {
		t.Fatalf("%d session rows, want 1", len(rows))
	}
	if rows[0].RevokedAt.IsZero() {
		t.Error("revoked_at was not set; the session was only forgotten by the browser")
	}

	// The cookie the browser was told to drop, sent anyway.
	if got := server.get(mePath, cookie).Code; got != http.StatusUnauthorized {
		t.Errorf("replaying the cookie after logout = %d, want 401", got)
	}
}

func TestLogoutWithoutASessionIsStillANoContent(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)

	if got := server.post(logoutPath, "").Code; got != http.StatusNoContent {
		t.Errorf("logout with no cookie = %d, want 204: the caller asked to be signed out and they are", got)
	}
}

// --- The current user ----------------------------------------------------------

func TestGetCurrentUserUnauthenticatedIsTheStandardProblem(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)

	recorder := server.get(mePath)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("GET /auth/me = %d, want 401", recorder.Code)
	}
	problem := decodeProblem(t, recorder)
	if problem.Code != gen.ProblemCodeUnauthenticated {
		t.Errorf("code = %q, want %q", problem.Code, gen.ProblemCodeUnauthenticated)
	}
	if problem.Instance == nil || *problem.Instance == "" {
		t.Error("the problem carries no request ID")
	}
}

func TestGetCurrentUserReportsMembershipsAndMFAState(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	if err := server.db.Write(t.Context(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.engagement (id, name, client, description, status, starts_on, ends_on, attack_version, mode, auto_reveal_on_start, created_by, created_at, updated_at)
			VALUES ('engagement-1', 'test', 'test', '', 'draft', '2026-01-01', '2026-06-01', '15.1', 'standard', false, 'u1', '2026-01-01 00:00:00', '2026-01-01 00:00:00')`,
		)
		return err
	}); err != nil {
		t.Fatalf("creating engagement: %v", err)
	}
	user := server.seedUser(t)
	if _, err := identity.NewMemberships(server.db).Add(t.Context(), identity.NewMembership{
		EngagementID: "engagement-1",
		UserID:       user.ID,
		Role:         authz.EngagementRoleBlue,
	}); err != nil {
		t.Fatalf("adding a membership: %v", err)
	}

	cookie := server.signIn(t)
	recorder := server.get(mePath, cookie)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET /auth/me = %d, want 200\nbody: %s", recorder.Code, recorder.Body)
	}

	current := decodeJSON[gen.CurrentUser](t, recorder)
	switch {
	case len(current.Memberships) != 1:
		t.Fatalf("%d memberships, want 1", len(current.Memberships))
	case current.Memberships[0].EngagementId != "engagement-1":
		t.Errorf("engagementId = %q, want %q", current.Memberships[0].EngagementId, "engagement-1")
	case current.Memberships[0].Role != gen.EngagementRoleBlue:
		t.Errorf("role = %q, want %q", current.Memberships[0].Role, gen.EngagementRoleBlue)
	case current.Mfa.Enforced:
		t.Error("mfa.enforced is true for a user nobody required it of")
	case current.Mfa.Satisfied:
		t.Error("mfa.satisfied is true for a session that presented no second factor")
	}
}

// TestTheRoleVocabulariesAgree: the roles are declared in api/openapi.yaml and
// again in internal/store/identity, and the handlers convert between them by
// casting. A value the database can hold that the API does not declare would be
// served as a string no client has a case for.
func TestTheRoleVocabulariesAgree(t *testing.T) {
	t.Parallel()

	platform := map[authz.PlatformRole]gen.PlatformRole{
		authz.PlatformRoleAdmin:  gen.PlatformRoleAdmin,
		authz.PlatformRoleMember: gen.PlatformRoleMember,
	}
	for stored, served := range platform {
		if string(stored) != string(served) {
			t.Errorf("the platform role %q is served as %q", stored, served)
		}
	}

	engagement := map[authz.EngagementRole]gen.EngagementRole{
		authz.EngagementRoleLead:     gen.EngagementRoleLead,
		authz.EngagementRoleRed:      gen.EngagementRoleRed,
		authz.EngagementRoleBlue:     gen.EngagementRoleBlue,
		authz.EngagementRoleObserver: gen.EngagementRoleObserver,
	}
	for stored, served := range engagement {
		if string(stored) != string(served) {
			t.Errorf("the engagement role %q is served as %q", stored, served)
		}
	}
}

// --- Sessions ending on their own -----------------------------------------------

func TestAnExpiredSessionIsRefusedAndIsNotReusable(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)
	cookie := server.signIn(t)

	server.execSQL(t, `UPDATE app.session SET expires_at = ?`, time.Now().Add(-time.Minute).UTC())

	if got := server.get(mePath, cookie).Code; got != http.StatusUnauthorized {
		t.Errorf("an expired session = %d, want 401", got)
	}
	// And again: nothing about the first refusal makes the second one succeed.
	if got := server.get(mePath, cookie).Code; got != http.StatusUnauthorized {
		t.Errorf("the second use of an expired session = %d, want 401", got)
	}
}

func TestAnIdleSessionIsRefused(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)
	cookie := server.signIn(t)

	// Well within its absolute expiry, but untouched for longer than the idle
	// timeout, so only the timeout can be what refuses it.
	server.execSQL(t, `UPDATE app.session SET last_seen_at = ?`, time.Now().Add(-3*time.Hour).UTC())

	if got := server.get(mePath, cookie).Code; got != http.StatusUnauthorized {
		t.Errorf("an idle session = %d, want 401", got)
	}
}

// TestDisablingAnAccountEndsItsSessionsNow: an administrator disabling somebody
// must not have to wait for their session to expire (M1-016 gives them the
// button; this is the half that makes it mean something).
func TestDisablingAnAccountEndsItsSessionsNow(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	user := server.seedUser(t)
	cookie := server.signIn(t)

	user.Status = identity.StatusDisabled
	if _, err := identity.NewUsers(server.db).Update(t.Context(), user); err != nil {
		t.Fatalf("disabling the account: %v", err)
	}

	if got := server.get(mePath, cookie).Code; got != http.StatusUnauthorized {
		t.Errorf("a disabled user's live session = %d, want 401", got)
	}
}

// --- Changing a password ---------------------------------------------------------

func TestChangingThePasswordRotatesThisSessionAndRevokesTheOthers(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)

	other := server.signIn(t)
	current := server.signIn(t)

	recorder := server.post(passwordPath,
		fmt.Sprintf(`{"currentPassword":%q,"newPassword":%q}`, testPassword, testNewPass),
		current)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("change password = %d, want 204\nbody: %s", recorder.Code, recorder.Body)
	}

	rotated := sessionCookie(t, recorder)
	if rotated.Value == current.Value {
		t.Fatal("the session token was not rotated")
	}
	if got := server.get(mePath, rotated).Code; got != http.StatusOK {
		t.Errorf("the rotated cookie = %d, want 200", got)
	}
	if got := server.get(mePath, current).Code; got != http.StatusUnauthorized {
		t.Errorf("the token that was rotated away = %d, want 401", got)
	}
	if got := server.get(mePath, other).Code; got != http.StatusUnauthorized {
		t.Errorf("another session of the same user = %d, want 401 — "+
			"changing a password signs out the places it was not changed from", got)
	}

	// And the new password is the one that works.
	if got := server.login(testEmail, testPassword).Code; got != http.StatusUnauthorized {
		t.Errorf("signing in with the old password = %d, want 401", got)
	}
	if got := server.login(testEmail, testNewPass).Code; got != http.StatusOK {
		t.Errorf("signing in with the new password = %d, want 200", got)
	}
}

func TestChangingThePasswordNeedsTheCurrentOne(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)
	cookie := server.signIn(t)

	recorder := server.post(passwordPath,
		fmt.Sprintf(`{"currentPassword":"not the current one","newPassword":%q}`, testNewPass),
		cookie)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("change password = %d, want 400\nbody: %s", recorder.Code, recorder.Body)
	}

	problem := decodeProblem(t, recorder)
	if problem.Errors == nil || len(*problem.Errors) == 0 {
		t.Fatalf("no field errors: %s", recorder.Body)
	}
	if got := (*problem.Errors)[0].Field; got != "currentPassword" {
		t.Errorf("the field error is on %q, want %q", got, "currentPassword")
	}

	// Nothing changed: the old password still signs in, the session still works.
	if got := server.get(mePath, cookie).Code; got != http.StatusOK {
		t.Errorf("the session = %d after a refused change, want 200", got)
	}
	if got := server.login(testEmail, testPassword).Code; got != http.StatusOK {
		t.Errorf("the old password = %d after a refused change, want 200", got)
	}
}

func TestANewPasswordIsHeldToThePolicy(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)
	cookie := server.signIn(t)

	tests := map[string]string{
		"too short":               "short",
		"one attackers try first": "password123456",
		"the one already in use":  testPassword,
	}
	for name, newPassword := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := server.post(passwordPath,
				fmt.Sprintf(`{"currentPassword":%q,"newPassword":%q}`, testPassword, newPassword),
				cookie)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("change password = %d, want 400\nbody: %s", recorder.Code, recorder.Body)
			}
			problem := decodeProblem(t, recorder)
			if problem.Errors == nil || len(*problem.Errors) == 0 {
				t.Fatalf("no field errors: %s", recorder.Body)
			}
			if got := (*problem.Errors)[0].Field; got != "newPassword" {
				t.Errorf("the field error is on %q, want %q", got, "newPassword")
			}
		})
	}
}

func TestChangingAPasswordNeedsASession(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)

	recorder := server.post(passwordPath,
		fmt.Sprintf(`{"currentPassword":%q,"newPassword":%q}`, testPassword, testNewPass))
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("change password with no session = %d, want 401\nbody: %s", recorder.Code, recorder.Body)
	}
}

// --- Hash upgrades and MFA ---------------------------------------------------------

// TestALoginUpgradesAHashMadeUnderWeakerSettings is M1-002's needsRehash,
// reaching an existing account through the front door.
func TestALoginUpgradesAHashMadeUnderWeakerSettings(t *testing.T) {
	t.Parallel()

	weak := password.Default()
	weak.Memory /= 4
	weak.Time = 1
	old, err := weak.Hash(testPassword)
	if err != nil {
		t.Fatalf("hashing under the old settings: %v", err)
	}

	server := newAuthServer(t)
	user := server.seedUser(t, func(u *identity.NewUser) { u.PasswordHash = old })

	if got := server.login(testEmail, testPassword).Code; got != http.StatusOK {
		t.Fatalf("login = %d, want 200 — an old hash still verifies", got)
	}

	reread, err := identity.NewUsers(server.db).ByID(t.Context(), user.ID)
	if err != nil {
		t.Fatalf("re-reading the user: %v", err)
	}
	if reread.PasswordHash == old {
		t.Fatal("the stored hash was not replaced")
	}
	stored, _, _, err := password.Decode(reread.PasswordHash)
	if err != nil {
		t.Fatalf("decoding the stored hash: %v", err)
	}
	if stored.Memory != password.Default().Memory || stored.Time != password.Default().Time {
		t.Errorf("the upgraded hash is m=%d,t=%d, want the current m=%d,t=%d",
			stored.Memory, stored.Time, password.Default().Memory, password.Default().Time)
	}

	// And the upgrade did not break the password.
	if got := server.login(testEmail, testPassword).Code; got != http.StatusOK {
		t.Errorf("login after the upgrade = %d, want 200", got)
	}
}

// What an account with MFA enforced and nothing enrolled gets is
// mfaenforcement_test.go's subject: before M1-008 it was a dead end, and now it
// is a session confined to enrolling.

// --- The request the specification does not describe ---------------------------------

// TestALoginBodyIsValidatedAgainstTheSpecification: the validator runs before
// the handler, so a body with a field that does not exist never reaches code
// that could act on it (PLAN.md §4).
func TestALoginBodyIsValidatedAgainstTheSpecification(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)
	server.seedUser(t)

	tests := map[string]string{
		"no password": `{"email":"alice@example.com"}`,
		"a field the specification does not have": `{"email":"alice@example.com","password":"x","platformRole":"admin"}`,
		"the wrong type": `{"email":"alice@example.com","password":42}`,
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := server.post(loginPath, body)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("login = %d, want 400\nbody: %s", recorder.Code, recorder.Body)
			}
			if got := decodeProblem(t, recorder).Code; got != gen.ProblemCodeValidationFailed {
				t.Errorf("code = %q, want %q", got, gen.ProblemCodeValidationFailed)
			}
		})
	}
}

// TestAnAuthenticatedEndpointIsReachableWithoutASession is about the middleware
// rather than the endpoint: authentication does not refuse anything, so a
// request with no cookie reaches the handler and the handler answers. M1-013
// moves that decision into one place; this is what it will change.
func TestAnAuthenticatedEndpointIsReachableWithoutASession(t *testing.T) {
	t.Parallel()

	server := newAuthServer(t)

	// 401 from the handler, not from the validator: the validator would have
	// answered before the request was routed, and the code would be different.
	recorder := server.get(mePath)
	if got := decodeProblem(t, recorder).Code; got != gen.ProblemCodeUnauthenticated {
		t.Errorf("code = %q, want %q", got, gen.ProblemCodeUnauthenticated)
	}
}
