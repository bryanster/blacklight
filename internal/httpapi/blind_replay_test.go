package httpapi

// BL-006 regression: the SSE catch-up replay path must re-derive blind
// visibility before delivery. Replayed payloads are rebuilt from stored
// activity rows and carry no `revealed` field, so the live-path filter
// (VisibleActivity) fails open for them; a blue seat reconnecting with a stale
// Last-Event-ID would otherwise receive step.created / execution.* /
// comment.created frames for unrevealed steps — the exact fact blind mode
// withholds — while the live stream drops the identical event.
//
// testConfig sets Events.MaxReplayEvents (500, the production default via
// BLACKLIGHT_EVENTS_MAX_REPLAY). Replay is silently skipped when the setting
// is 0, so a test config that leaves it at zero exercises nothing — the
// harness never ran this path before BL-006 for exactly that reason.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/config"
	"github.com/bryanster/blacklight/internal/events"
)

// staleCursor is a zero UUIDv7: it sorts before every real activity id, so
// replaying from it delivers every recorded event for the engagement.
const staleCursor = "00000000-0000-7000-8000-000000000000"

func captureReplay(t *testing.T, ts *httptest.Server, cookie *http.Cookie, engID string) []sseFrame {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	frames := captureSSE(t,
		ts.URL+eventsPathTest+"?topics="+events.EngagementTopic(engID)+"&lastEventId="+staleCursor,
		cookie, ctx)
	// Replay is written by the streamWithReplay goroutine right after the
	// handshake; give it a moment, then drain whatever has queued.
	time.Sleep(400 * time.Millisecond)
	return drainFrames(t, frames, 500*time.Millisecond)
}

func replayedTypes(t *testing.T, frames []sseFrame) []string {
	t.Helper()
	types := make([]string, 0, len(frames))
	for _, fr := range frames {
		types = append(types, parseEnvelope(t, fr).Type)
	}
	return types
}

// seedBlindActivity drives the real handlers as red, so the activity rows are
// the ones production writes: step.created, step.updated,
// execution.red_updated, comment.created — all about one still-unrevealed
// step. Returns the step and execution ids the events name.
func seedBlindActivity(t *testing.T, server *authServer, engID, scenarioID string, red *http.Cookie) (stepID, execID string) {
	t.Helper()

	rec := server.post(fmt.Sprintf("%s/engagements/%s/scenarios/%s/steps", BasePath, engID, scenarioID),
		`{"name":"hidden-step"}`, red)
	if rec.Code != http.StatusCreated {
		t.Fatalf("red create step = %d\n%s", rec.Code, rec.Body)
	}
	var step struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &step); err != nil {
		t.Fatal(err)
	}

	rec = server.send(http.MethodPatch,
		fmt.Sprintf("%s/engagements/%s/scenarios/%s/steps/%s", BasePath, engID, scenarioID, step.ID),
		`{"name":"renamed"}`, red)
	if rec.Code != http.StatusOK {
		t.Fatalf("red patch step = %d\n%s", rec.Code, rec.Body)
	}

	var execs struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	rec = server.get(fmt.Sprintf("%s/engagements/%s/executions?scenarioId=%s", BasePath, engID, scenarioID), red)
	if rec.Code != http.StatusOK {
		t.Fatalf("red list executions = %d\n%s", rec.Code, rec.Body)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &execs); err != nil {
		t.Fatal(err)
	}
	if len(execs.Items) == 0 {
		t.Fatal("no execution for the created step")
	}
	execID = execs.Items[0].ID

	rec = server.send(http.MethodPatch,
		fmt.Sprintf("%s/engagements/%s/executions/%s/execution", BasePath, engID, execID),
		`{"version":1,"status":"running"}`, red)
	if rec.Code != http.StatusOK {
		t.Fatalf("red patch execution = %d\n%s", rec.Code, rec.Body)
	}

	rec = server.post(fmt.Sprintf("%s/engagements/%s/executions/%s/comments", BasePath, engID, execID),
		`{"body":"war room note"}`, red)
	if rec.Code != http.StatusCreated {
		t.Fatalf("red create comment = %d\n%s", rec.Code, rec.Body)
	}
	return step.ID, execID
}

func raiseFinding(t *testing.T, server *authServer, engID string, who *http.Cookie) {
	t.Helper()
	rec := server.post(fmt.Sprintf("%s/engagements/%s/findings", BasePath, engID),
		`{"title":"leak check","description":"non-step-scoped replay control","severity":"low"}`, who)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create finding = %d\n%s", rec.Code, rec.Body)
	}
}

// TestBlindReplayWithholdsUnrevealedStepEvents is the ticket's proof, inverted
// into the fixed behaviour: every request from the reproduction, asserting
// absence of the hidden step instead of its presence.
func TestBlindReplayWithholdsUnrevealedStepEvents(t *testing.T) {
	t.Parallel()
	server := newAuthServer(t)

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")

	engID := "01900000-b006-7000-8000-000000000001"
	scenarioID := "01900000-b006-7000-8000-000000000002"

	seedBlindEngagementDB(t, server, engID, scenarioID, red, blue)

	redCookie := sessionCookie(t, server.login(red.Email, testPassword))
	blueCookie := sessionCookie(t, server.login(blue.Email, testPassword))

	stepID, execID := seedBlindActivity(t, server, engID, scenarioID, redCookie)
	raiseFinding(t, server, engID, redCookie)

	// The middleware blind guard holds: blue's direct read of the step is 404.
	rec := server.get(fmt.Sprintf("%s/engagements/%s/scenarios/%s/steps/%s", BasePath, engID, scenarioID, stepID), blueCookie)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("blue fetching the unrevealed step = %d, want 404\n%s", rec.Code, rec.Body)
	}

	ts := httptest.NewServer(server.handler)
	t.Cleanup(ts.Close)

	// Blue reconnects with a stale cursor.
	frames := captureReplay(t, ts, blueCookie, engID)

	// No replayed frame may name the hidden step — directly (step.*) or via
	// parent linkage (execution.*, comment.*). The finding event is
	// non-step-scoped and must still arrive: blind mode withholds the step
	// links, not the finding itself.
	var findingReplayed bool
	for _, fr := range frames {
		env := parseEnvelope(t, fr)
		if env.Type == "finding.created" {
			findingReplayed = true
		}
		if strings.Contains(string(env.Data), stepID) || strings.Contains(string(env.Data), execID) {
			t.Errorf("blue replay received %q naming an unrevealed step or its execution: %s", env.Type, fr.data)
		}
	}
	if !findingReplayed {
		t.Errorf("blue replay lost the non-step-scoped finding.created event; got %v", replayedTypes(t, frames))
	}
	if len(frames) == 0 {
		t.Fatal("blue replay delivered nothing at all — replay is disabled or the filter over-withholds")
	}

	// The same cursor after reveal: reveal state is re-derived at delivery
	// time (live lookup, not a write-time snapshot), so the very events that
	// were withheld now arrive.
	rec = server.post(fmt.Sprintf("%s/engagements/%s/scenarios/%s/steps/%s/reveal", BasePath, engID, scenarioID, stepID),
		"", redCookie)
	if rec.Code != http.StatusOK {
		t.Fatalf("red reveal step = %d\n%s", rec.Code, rec.Body)
	}

	frames = captureReplay(t, ts, blueCookie, engID)
	var gotStepCreated, gotExecUpdated, gotStepRevealed bool
	for _, fr := range frames {
		env := parseEnvelope(t, fr)
		switch env.Type {
		case "step.created":
			gotStepCreated = true
		case "execution.red_updated":
			gotExecUpdated = true
		case "step.revealed":
			gotStepRevealed = true
		}
		if strings.Contains(string(env.Data), stepID) {
			var d events.EventData
			if err := json.Unmarshal(env.Data, &d); err != nil {
				t.Fatalf("replayed envelope unparseable: %v\n%s", err, fr.data)
			}
			if d.Revealed == nil || !*d.Revealed {
				t.Errorf("blue replay received %q without revealed=true after reveal: %s", env.Type, fr.data)
			}
		}
	}
	if !gotStepCreated || !gotExecUpdated || !gotStepRevealed {
		t.Errorf("blue replay after reveal lost events (step.created=%v execution.red_updated=%v step.revealed=%v); got %v",
			gotStepCreated, gotExecUpdated, gotStepRevealed, replayedTypes(t, frames))
	}
}

// TestBlindReplayRedUnchanged pins the other acceptance criterion: red/lead
// replay behaviour is untouched by the filter.
func TestBlindReplayRedUnchanged(t *testing.T) {
	t.Parallel()
	server := newAuthServer(t)

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")
	engID := "01900000-b006-7000-8000-000000000011"
	scenarioID := "01900000-b006-7000-8000-000000000012"

	seedBlindEngagementDB(t, server, engID, scenarioID, red, blue)

	redCookie := sessionCookie(t, server.login(red.Email, testPassword))
	stepID, _ := seedBlindActivity(t, server, engID, scenarioID, redCookie)

	ts := httptest.NewServer(server.handler)
	t.Cleanup(ts.Close)

	frames := captureReplay(t, ts, redCookie, engID)
	types := replayedTypes(t, frames)
	for _, want := range []string{"step.created", "step.updated", "execution.red_updated", "comment.created"} {
		found := false
		for _, got := range types {
			if got == want {
				found = true
			}
		}
		if !found {
			t.Errorf("red replay lost %q; got %v", want, types)
		}
	}
	_ = stepID
}

// TestReplayTruncationEmitsGap covers the truncation contract the fix must not
// disturb: more rows than MaxReplayEvents replays the cap and emits stream.gap.
func TestReplayTruncationEmitsGap(t *testing.T) {
	t.Parallel()
	server := newAuthServer(t, func(cfg *config.Config) { cfg.Events.MaxReplayEvents = 1 })

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")
	engID := "01900000-b006-7000-8000-000000000021"
	scenarioID := "01900000-b006-7000-8000-000000000022"

	seedBlindEngagementDB(t, server, engID, scenarioID, red, blue)

	redCookie := sessionCookie(t, server.login(red.Email, testPassword))
	seedBlindActivity(t, server, engID, scenarioID, redCookie)

	ts := httptest.NewServer(server.handler)
	t.Cleanup(ts.Close)

	frames := captureReplay(t, ts, redCookie, engID)
	gaps := 0
	for _, fr := range frames {
		if parseEnvelope(t, fr).Type == events.TypeStreamGap {
			gaps++
		}
	}
	if gaps != 1 {
		t.Errorf("stream.gap count = %d, want 1 (replay truncated); frames: %v", gaps, replayedTypes(t, frames))
	}
}

// TestReplayDisabledWhenMaxReplayEventsZero pins the off switch: 0 disables
// replay entirely — only live tail runs.
func TestReplayDisabledWhenMaxReplayEventsZero(t *testing.T) {
	t.Parallel()
	server := newAuthServer(t, func(cfg *config.Config) { cfg.Events.MaxReplayEvents = 0 })

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")
	engID := "01900000-b006-7000-8000-000000000021"
	scenarioID := "01900000-b006-7000-8000-000000000022"

	seedBlindEngagementDB(t, server, engID, scenarioID, red, blue)

	redCookie := sessionCookie(t, server.login(red.Email, testPassword))
	seedBlindActivity(t, server, engID, scenarioID, redCookie)

	ts := httptest.NewServer(server.handler)
	t.Cleanup(ts.Close)

	frames := captureReplay(t, ts, redCookie, engID)
	for _, fr := range frames {
		if parseEnvelope(t, fr).Type == "step.created" {
			t.Fatal("step.created replayed although MaxReplayEvents is 0")
		}
	}
}
