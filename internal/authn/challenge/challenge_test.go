package challenge_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/authn/challenge"
	"github.com/bryanster/blacklight/internal/authz"
	"github.com/bryanster/blacklight/internal/config"
	"github.com/bryanster/blacklight/internal/store/identity"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

// unusedStore is a Store whose methods must never be reached. New's option
// checks are startup checks — they run before anything touches the store — so a
// test can prove a bad configuration is rejected without a database.
type unusedStore struct{}

func (unusedStore) Open(context.Context, identity.NewMFAChallenge) (identity.MFAChallenge, error) {
	panic("challenge: Open reached on an invalid configuration")
}

func (unusedStore) ByTokenHash(context.Context, string) (identity.MFAChallenge, error) {
	panic("challenge: ByTokenHash reached on an invalid configuration")
}

func (unusedStore) Consume(context.Context, string, time.Time) (bool, error) {
	panic("challenge: Consume reached on an invalid configuration")
}

func (unusedStore) DeleteForUser(context.Context, string) error {
	panic("challenge: DeleteForUser reached on an invalid configuration")
}

func validOptions() challenge.Options {
	return challenge.Options{
		Secret: []byte("a-session-secret"),
		TTL:    5 * time.Minute,
		Path:   "/api/v1/auth/mfa",
	}
}

func TestNewRejectsAConfigurationThatCannotIssueAChallenge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		opts challenge.Options
		want string
	}{
		{"no secret", func() challenge.Options { o := validOptions(); o.Secret = nil; return o }(), "no secret"},
		{"zero window", func() challenge.Options { o := validOptions(); o.TTL = 0; return o }(), "positive"},
		{"negative window", func() challenge.Options { o := validOptions(); o.TTL = -time.Second; return o }(), "positive"},
		{"relative path", func() challenge.Options { o := validOptions(); o.Path = "auth/mfa"; return o }(), "absolute"},
		{"empty path", func() challenge.Options { o := validOptions(); o.Path = ""; return o }(), "absolute"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := challenge.New(unusedStore{}, tc.opts); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want one containing %q", err, tc.want)
			}
		})
	}

	// No store is its own refusal: there would be nowhere to keep a challenge.
	if _, err := challenge.New(nil, validOptions()); err == nil || !strings.Contains(err.Error(), "no store") {
		t.Fatalf("New(nil store) err = %v, want a no-store refusal", err)
	}

	// And the valid configuration is accepted.
	if _, err := challenge.New(unusedStore{}, validOptions()); err != nil {
		t.Fatalf("New(valid) = %v, want nil", err)
	}
}

func TestOptionsFromDerivesTheCookiePolicyFromTheDeployment(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Env:     config.EnvProduction,
		Session: config.Session{Secret: config.NewSecret([]byte("the-session-secret"))},
		MFA:     config.MFA{PendingTTL: 90 * time.Second},
	}

	opts := challenge.OptionsFrom(cfg, "/api/v1/auth/mfa")
	if string(opts.Secret) != "the-session-secret" {
		t.Errorf("secret = %q, want the session secret", opts.Secret)
	}
	if opts.TTL != 90*time.Second {
		t.Errorf("ttl = %v, want the configured pending window", opts.TTL)
	}
	if !opts.Secure {
		t.Error("Secure = false in production, want true")
	}
	if opts.Path != "/api/v1/auth/mfa" {
		t.Errorf("path = %q, want the MFA prefix it was given", opts.Path)
	}

	// Development is the one posture that relaxes Secure, for an http host.
	cfg.Env = config.EnvDevelopment
	if challenge.OptionsFrom(cfg, "/x").Secure {
		t.Error("Secure = true in development, want the documented relaxation")
	}
}

// --- The token as a credential --------------------------------------------------

func TestATokenIsRedactedWhereverItIsRendered(t *testing.T) {
	t.Parallel()

	const secret = "the-real-token-value"
	tok := challenge.Token(secret)

	renderings := map[string]string{
		"String":        tok.String(),
		"GoString":      tok.GoString(),
		"string verb":   fmt.Sprintf("%s", tok),
		"value verb":    fmt.Sprintf("%v", tok),
		"go-syntax":     fmt.Sprintf("%#v", tok),
		"untyped verb":  fmt.Sprintf("%d", tok),
		"quoted string": fmt.Sprintf("%q", tok),
	}
	for name, got := range renderings {
		if strings.Contains(got, secret) {
			t.Errorf("%s rendered the token value: %q", name, got)
		}
	}
	// Every verb but %q yields the bare placeholder; %q quotes it.
	for name, got := range renderings {
		want := "[redacted]"
		if name == "quoted string" {
			want = `"[redacted]"`
		}
		if got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}

	// Reveal is the one way to the value, and it exists for the Set-Cookie
	// header and the hash.
	if got := tok.Reveal(); got != secret {
		t.Errorf("Reveal() = %q, want the token", got)
	}
}

func TestATokenNeverMarshalsItsValue(t *testing.T) {
	t.Parallel()

	const secret = "the-real-token-value"
	tok := challenge.Token(secret)

	encoded, err := json.Marshal(tok)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(encoded) != `"[redacted]"` {
		t.Fatalf("json = %s, want [redacted]", encoded)
	}

	text, err := tok.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	if string(text) != "[redacted]" {
		t.Fatalf("text = %s, want [redacted]", text)
	}

	// The same holds as a struct field, which is how a token would actually
	// reach a response body.
	wrapped, err := json.Marshal(struct {
		Token challenge.Token `json:"token"`
	}{tok})
	if err != nil {
		t.Fatalf("Marshal(struct): %v", err)
	}
	if strings.Contains(string(wrapped), secret) {
		t.Fatalf("a token field serialized its value: %s", wrapped)
	}
}

func TestATokenLoggedIsRedacted(t *testing.T) {
	t.Parallel()

	const secret = "the-real-token-value"
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	logger.Info("pending second factor", "token", challenge.Token(secret))

	if strings.Contains(buf.String(), secret) {
		t.Fatalf("the token reached the log: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "[redacted]") {
		t.Fatalf("log line = %s, want the redaction placeholder", buf.String())
	}
}

// --- The cookie -----------------------------------------------------------------

func TestCookieCarriesThePendingTokenAndItsProtections(t *testing.T) {
	t.Parallel()

	m, err := challenge.New(unusedStore{}, challenge.Options{
		Secret: []byte("k"),
		TTL:    5 * time.Minute,
		Secure: true,
		Path:   "/api/v1/auth/mfa",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	expires := time.Now().Add(time.Hour)
	tok := challenge.Token("pending-token")
	c := m.Cookie(tok, expires)

	if c.Name != challenge.CookieName {
		t.Errorf("name = %q, want %q", c.Name, challenge.CookieName)
	}
	if c.Value != tok.Reveal() {
		t.Errorf("value = %q, want the token itself", c.Value)
	}
	if c.Path != "/api/v1/auth/mfa" {
		t.Errorf("path = %q, want the MFA prefix", c.Path)
	}
	if !c.HttpOnly || !c.Secure || c.SameSite != http.SameSiteStrictMode {
		t.Errorf("attributes = HttpOnly:%v Secure:%v SameSite:%v, want all protections on", c.HttpOnly, c.Secure, c.SameSite)
	}
	if !c.Expires.Equal(expires.UTC()) {
		t.Errorf("expires = %v, want %v", c.Expires, expires.UTC())
	}
	if c.MaxAge <= 0 {
		t.Errorf("MaxAge = %d, want a positive lifetime", c.MaxAge)
	}

	// The clear has to match the set or a browser keeps the original.
	clear := m.ClearCookie()
	if clear.Name != c.Name || clear.Path != c.Path || clear.HttpOnly != c.HttpOnly ||
		clear.Secure != c.Secure || clear.SameSite != c.SameSite {
		t.Fatalf("clear attributes %+v do not match the set attributes %+v", clear, c)
	}
	if clear.Value != "" || clear.MaxAge != -1 {
		t.Errorf("clear = value %q maxAge %d, want an empty value and MaxAge -1", clear.Value, clear.MaxAge)
	}
	if !clear.Expires.Before(time.Now()) {
		t.Errorf("clear expires = %v, want a time in the past", clear.Expires)
	}
}

func TestFromRequestReadsOnlyThePendingCookie(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/mfa/totp/verify", nil)
	if got := challenge.FromRequest(r); got != "" {
		t.Errorf("no cookie: got %q, want the empty token", got)
	}

	// A session cookie is not a pending token, whatever it holds.
	r.AddCookie(&http.Cookie{Name: "bl_session", Value: "a-session-token"})
	if got := challenge.FromRequest(r); got != "" {
		t.Errorf("session cookie read as pending: %q", got)
	}

	r.AddCookie(&http.Cookie{Name: challenge.CookieName, Value: "a-pending-token"})
	if got := challenge.FromRequest(r); got.Reveal() != "a-pending-token" {
		t.Errorf("pending cookie = %q, want its value", got.Reveal())
	}
}

// --- Against the real store -----------------------------------------------------
//
// The manager's own rules — expiry, single use, supersede, and one error for
// every "nothing pending here" — driven through the identity store, with the
// clock held still so expiry is a fact and not a wait.

func TestResolveSpendAndDiscardAgainstTheRealStore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := storetest.Migrated(t)
	users := identity.NewUsers(db)
	newUser := func(email string) string {
		t.Helper()
		u, err := users.Create(ctx, identity.NewUser{
			Email:        email,
			DisplayName:  email,
			PasswordHash: "argon2id$placeholder",
			PlatformRole: authz.PlatformRoleMember,
			Status:       identity.StatusActive,
		})
		if err != nil {
			t.Fatalf("create user %q: %v", email, err)
		}
		return u.ID
	}
	user1, user2, user3, user4 := newUser("one@example.test"), newUser("two@example.test"), newUser("three@example.test"), newUser("four@example.test")

	now := time.Unix(1_700_000_000, 0)
	m, err := challenge.New(identity.NewMFAChallenges(db), challenge.Options{
		Secret: []byte("k"),
		TTL:    time.Minute,
		Path:   "/api/v1/auth/mfa",
		Now:    func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Open mints a challenge that resolves to its user, and stamps the expiry
	// from the manager's clock and window.
	issued, err := m.Open(ctx, user1)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if issued.Token.Reveal() == "" {
		t.Fatal("Open returned an empty token")
	}
	if want := now.Add(time.Minute); !issued.Challenge.ExpiresAt.Equal(want) {
		t.Errorf("expires_at = %v, want %v", issued.Challenge.ExpiresAt, want)
	}
	found, err := m.Resolve(ctx, issued.Token)
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if found.UserID != user1 {
		t.Errorf("resolved user = %q, want %q", found.UserID, user1)
	}

	// A token of the wrong length is refused without a lookup, and so is a
	// well-formed token that resolves to nothing.
	if _, err := m.Resolve(ctx, "short"); !errors.Is(err, challenge.ErrNoChallenge) {
		t.Errorf("short token: err = %v, want ErrNoChallenge", err)
	}
	orphan, err := m.Open(ctx, user2)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := m.Discard(ctx, user2); err != nil {
		t.Fatalf("Discard: %v", err)
	}
	if _, err := m.Resolve(ctx, orphan.Token); !errors.Is(err, challenge.ErrNoChallenge) {
		t.Errorf("discarded token: err = %v, want ErrNoChallenge", err)
	}

	// One correct use spends the challenge exactly once.
	if _, err := m.Resolve(ctx, issued.Token); err != nil {
		t.Fatalf("Resolve before spend: %v", err)
	}
	spent, err := m.Spend(ctx, issued.Challenge.ID)
	if err != nil {
		t.Fatalf("Spend: %v", err)
	}
	if !spent {
		t.Fatal("the first Spend reported the challenge already spent")
	}
	if _, err := m.Resolve(ctx, issued.Token); !errors.Is(err, challenge.ErrNoChallenge) {
		t.Errorf("spent token: err = %v, want ErrNoChallenge", err)
	}
	again, err := m.Spend(ctx, issued.Challenge.ID)
	if err != nil {
		t.Fatalf("second Spend: %v", err)
	}
	if again {
		t.Error("the second Spend reported it spent the challenge")
	}

	// The window is enforced: past the TTL the challenge resolves to nothing,
	// whatever the browser still holds.
	next, err := m.Open(ctx, user3)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	now = now.Add(2 * time.Minute)
	if _, err := m.Resolve(ctx, next.Token); !errors.Is(err, challenge.ErrNoChallenge) {
		t.Errorf("expired token: err = %v, want ErrNoChallenge", err)
	}

	// Opening again supersedes the previous challenge for that user.
	now = now.Add(time.Hour)
	first, err := m.Open(ctx, user4)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	second, err := m.Open(ctx, user4)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	if _, err := m.Resolve(ctx, first.Token); !errors.Is(err, challenge.ErrNoChallenge) {
		t.Errorf("superseded token: err = %v, want ErrNoChallenge", err)
	}
	if _, err := m.Resolve(ctx, second.Token); err != nil {
		t.Errorf("current token after supersede: %v", err)
	}
}
