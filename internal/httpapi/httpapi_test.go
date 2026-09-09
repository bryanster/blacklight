package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/config"
	"github.com/bryanster/blacklight/internal/httpapi/gen"
	"github.com/bryanster/blacklight/internal/store"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

// The helpers every test in this package shares: a server built the way the
// process builds it, a place for its log to go, and the two ways of reading a
// response body that the assertions need.

// testBaseURL is where a browser reaches the server under test. It is a
// constant because the SAML tests (M1-010) build an assertion consumer URL from
// it and mint assertions addressed to that exact string — the identity provider
// and this server have to agree on it byte for byte.
const testBaseURL = "http://localhost:8080"

// testConfig is a plausible production configuration. Tests that care about a
// particular value — an https base URL, a short request timeout — copy it and
// change that one field, so the rest stays realistic.
func testConfig(t *testing.T) config.Config {
	t.Helper()

	var baseURL config.URL
	if err := baseURL.UnmarshalText([]byte(testBaseURL)); err != nil {
		t.Fatalf("parsing the test base URL: %v", err)
	}
	return config.Config{
		Env: config.EnvProduction,
		Server: config.Server{
			Addr:            "127.0.0.1:0",
			BaseURL:         baseURL,
			RequestTimeout:  30 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		// A server cannot be built without these: it would have no way to hash a
		// session token. The secret is a fixed test value on purpose — nothing
		// here depends on it being unguessable, and a random one would make a
		// failure impossible to reproduce.
		Session: config.Session{
			Secret:      config.NewSecret([]byte("test-session-secret-not-a-real-one")),
			Lifetime:    12 * time.Hour,
			IdleTimeout: 2 * time.Hour,
		},
		// Distinct from the session secret, as config insists a real deployment
		// keeps them: it is what TOTP enrolments are encrypted under (M1-006).
		Encryption: config.Encryption{
			Key: config.NewSecret([]byte("test-encryption-key-also-not-real")),
		},
		MFA: config.MFA{PendingTTL: 5 * time.Minute},
		// The documented defaults. A server cannot be built without them either
		// — a zero threshold would lock out the first caller through the door —
		// and the tests that are about throttling lower them.
		Throttle: config.Throttle{
			AccountFailures: 5,
			AccountLockout:  15 * time.Minute,
			SourceFailures:  50,
			SourceLockout:   15 * time.Minute,
		},
		Content: config.Content{
			Dir:        t.TempDir(),
			MaxBytes:   config.ByteSize(512 << 20),
			JobTimeout: 30 * time.Minute,
			WriteBatch: 250,
		},
		Events: config.Events{
			MaxSubscribers: 256,
			Buffer:         16,
			Heartbeat:      15 * time.Second,
			// Replay must be on or Last-Event-ID tests exercise nothing: 0
			// silently disables the catch-up path (BL-006), and production
			// defaults it to 500 via BLACKLIGHT_EVENTS_MAX_REPLAY.
			MaxReplayEvents: 500,
		},
	}
}

// newTestServer builds the real chain over a real, empty database.
func newTestServer(t *testing.T) (http.Handler, *logBuffer) {
	t.Helper()
	return newTestServerWith(t, testConfig(t), storetest.New(t))
}

func newTestServerWith(t *testing.T, cfg config.Config, db store.Store, adjust ...func(*Deps)) (http.Handler, *logBuffer) {
	t.Helper()

	logs := &logBuffer{}
	deps := Deps{Config: cfg, Store: db, Logger: logs.logger()}
	// Incomplete test doubles (panickyStore, stubStore, deadlineStore, …)
	// cannot answer content boot writes. Only a real *store.DB runs the worker.
	if _, ok := db.(*store.DB); !ok {
		deps.DisableContentRunner = true
	}
	for _, fn := range adjust {
		fn(&deps)
	}
	server, err := NewServer(deps)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return server, logs
}

// get performs a request against a handler and returns the recorded response.
func do(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func get(handler http.Handler, target string) *httptest.ResponseRecorder {
	return do(handler, httptest.NewRequest(http.MethodGet, target, nil))
}

// decodeProblem reads a response as a problem document, failing the test if it
// is not one — including the media type, which is half of what makes it a
// problem document rather than some JSON that happens to have a status in it.
func decodeProblem(t *testing.T, recorder *httptest.ResponseRecorder) gen.Problem {
	t.Helper()

	if got, want := recorder.Header().Get("Content-Type"), "application/problem+json"; got != want {
		t.Errorf("Content-Type = %q, want %q\nbody: %s", got, want, recorder.Body.String())
	}

	var problem gen.Problem
	if err := json.Unmarshal(recorder.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decoding the response as a problem document: %v\nbody: %s", err, recorder.Body.String())
	}
	if problem.Status != recorder.Code {
		t.Errorf("problem.status = %d but the response status is %d; a forwarded body would disagree with itself",
			problem.Status, recorder.Code)
	}
	return problem
}

func decodeJSON[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()

	var value T
	if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
		t.Fatalf("decoding the response: %v\nbody: %s", err, recorder.Body.String())
	}
	return value
}

// logBuffer collects what the server logged. The mutex is not decoration: the
// shutdown test reads it while a request is still being served.
type logBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (l *logBuffer) logger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(l, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func (l *logBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.Write(p)
}

func (l *logBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.String()
}

// lines returns the log as decoded records, so an assertion can name a field
// rather than matching a substring of the whole file.
func (l *logBuffer) lines(t *testing.T) []map[string]any {
	t.Helper()

	var records []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(l.String()), "\n") {
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("the log is not JSON: %v\nline: %s", err, line)
		}
		records = append(records, record)
	}
	return records
}

// find returns the first log record with the given message.
func (l *logBuffer) find(t *testing.T, message string) map[string]any {
	t.Helper()

	for _, record := range l.lines(t) {
		if record["msg"] == message {
			return record
		}
	}
	t.Fatalf("no %q line in the log:\n%s", message, l.String())
	return nil
}

// panickyStore is a store whose health check panics, which is how the panic
// path is reached through the real chain: no handler in this milestone panics
// on purpose, and a test that called the middleware directly would not prove
// the chain is wired the way it is.
type panickyStore struct {
	store.Store
}

func (panickyStore) Health(context.Context) error {
	panic("the database driver exploded: dsn=file:/secret/blacklight.duckdb")
}

// stubStore is the minimum a Deps needs. Embedding the interface means a method
// nobody calls does not need writing, and calling one is a nil-pointer panic in
// the test rather than a silently wrong answer.
type stubStore struct {
	store.Store
}

func (stubStore) Health(context.Context) error { return nil }

func (stubStore) Read() store.Reader { return nil }

func (stubStore) Write(context.Context, func(*sql.Tx) error) error { return nil }

func (stubStore) Close() error { return nil }

// stubContentLookupIP resolves every hostname to a public address, so content
// URL validation never touches the network in tests. Tests that exercise the
// SSRF fence (M7-014) override it via Deps.ContentLookupIP.
func stubContentLookupIP(_ context.Context, _ string) ([]net.IP, error) {
	return []net.IP{net.ParseIP("8.8.8.8")}, nil
}
