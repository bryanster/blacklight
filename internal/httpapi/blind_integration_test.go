package httpapi

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/bryanster/blacklight/internal/authz"
	"github.com/bryanster/blacklight/internal/events"
	"github.com/bryanster/blacklight/internal/events/presence"
	"github.com/bryanster/blacklight/internal/httpapi/gen"
	"github.com/bryanster/blacklight/internal/store/blind"
	"github.com/bryanster/blacklight/internal/store/identity"
)

// ---------------------------------------------------------------------------
// SSE helpers
// ---------------------------------------------------------------------------

type sseFrame struct {
	event string
	data  string
}

type sseEnvelope struct {
	ID    string          `json:"id"`
	Topic string          `json:"topic"`
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data"`
}

func captureSSE(t *testing.T, url string, cookie *http.Cookie, ctx context.Context) <-chan sseFrame {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("SSE connect: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if err != nil {
			t.Fatalf("SSE status = %d (and body read failed: %v)", resp.StatusCode, err)
		}
		t.Fatalf("SSE status = %d\n%s", resp.StatusCode, body)
	}

	frames := make(chan sseFrame, 64)
	go func() {
		defer close(frames)
		sc := bufio.NewScanner(resp.Body)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		var cur sseFrame
		for sc.Scan() {
			line := sc.Text()
			switch {
			case line == "":
				if cur.event != "" || cur.data != "" {
					frames <- cur
					cur = sseFrame{}
				}
			case strings.HasPrefix(line, ":"):
			case strings.HasPrefix(line, "event:"):
				cur.event = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:"):
				cur.data = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			}
		}
	}()
	return frames
}

func parseEnvelope(t *testing.T, fr sseFrame) sseEnvelope {
	t.Helper()
	var env sseEnvelope
	if err := json.Unmarshal([]byte(fr.data), &env); err != nil {
		t.Fatalf("envelope parse: %v\n%s", err, fr.data)
	}
	return env
}

func drainFrames(t *testing.T, frames <-chan sseFrame, timeout time.Duration) []sseFrame {
	t.Helper()
	var collected []sseFrame
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			return collected
		case fr, ok := <-frames:
			if !ok {
				return collected
			}
			collected = append(collected, fr)
		}
	}
}

func waitForFrame(t *testing.T, frames <-chan sseFrame, wantType string, timeout time.Duration) sseFrame {
	t.Helper()
	deadline := time.After(timeout)
	for {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for SSE frame type %q", wantType)
		case fr, ok := <-frames:
			if !ok {
				t.Fatalf("stream closed waiting for SSE frame type %q", wantType)
			}
			env := parseEnvelope(t, fr)
			if env.Type == wantType {
				return fr
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Seed helpers
// ---------------------------------------------------------------------------

func seedBlindEngagementDB(t *testing.T, server *authServer,
	engID, scenarioID string, red identity.User, blue identity.User,
) {
	t.Helper()
	if err := server.db.Write(t.Context(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.engagement (id, name, client, description, status,
			 starts_on, ends_on, attack_version, mode, auto_reveal_on_start,
			 created_by, created_at, updated_at)
			VALUES (?, 'blind-test', 'test', '', 'active',
			 '2026-01-01', '2026-06-01', '15.1', 'blind', false,
			 ?, '2026-01-01 00:00:00', '2026-01-01 00:00:00')`,
			engID, red.ID,
		)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(t.Context(),
			`INSERT INTO app.scenario (id, engagement_id, ordinal, name, narrative,
			 source, threat_actor, source_ref, plan_id, created_at, updated_at)
			VALUES (?, ?, 0, 'scenario-1', '',
			 'manual', '', '', '', '2026-01-01 00:00:00', '2026-01-01 00:00:00')`,
			scenarioID, engID,
		)
		return err
	}); err != nil {
		t.Fatalf("seeding blind engagement: %v", err)
	}

	memberships := identity.NewMemberships(server.db)
	if _, err := memberships.Add(t.Context(), identity.NewMembership{
		EngagementID: engID, UserID: red.ID, Role: authz.EngagementRoleRed,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := memberships.Add(t.Context(), identity.NewMembership{
		EngagementID: engID, UserID: blue.ID, Role: authz.EngagementRoleBlue,
	}); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Test 1: Presence focus stripped for blue in blind mode (REST)
// ---------------------------------------------------------------------------

func TestBlindPresenceFocusStrippedForBlueREST(t *testing.T) {
	t.Parallel()
	server := newAuthServerDeps(t, func(d *Deps) {
		d.Presence = presence.New(presence.Options{
			HeartbeatTTL: 60 * time.Second,
		})
	})

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")

	engID := "01900000-b001-7000-8000-000000000001"
	scenarioID := "01900000-b001-7000-8000-000000000002"
	stepID := "01900000-b001-7000-8000-000000000003"

	seedBlindEngagementDB(t, server, engID, scenarioID, red, blue)
	seedUnrevealedStep(t, server, stepID, scenarioID)

	// Red sets presence with focus on unrevealed step.
	redCookie := sessionCookie(t, server.login(red.Email, testPassword))
	presenceUUID := uuid.Must(uuid.NewV7())
	stepUUID := uuid.MustParse(stepID)
	setPresence(t, server, engID, presenceUUID, &stepUUID, nil, redCookie)

	// Blue sees red online but focus stripped.
	blueCookie := sessionCookie(t, server.login(blue.Email, testPassword))
	assertFocusStripped(t, server, engID, blueCookie, false)

	// Red sees focus intact.
	assertFocusStripped(t, server, engID, redCookie, true)
}

// ---------------------------------------------------------------------------
// Test 2: Presence focus stripped for blue in blind mode (SSE)
// ---------------------------------------------------------------------------

func TestBlindPresenceFocusStrippedForBlueSSE(t *testing.T) {
	t.Parallel()
	server := newAuthServerDeps(t, func(d *Deps) {
		d.Presence = presence.New(presence.Options{
			HeartbeatTTL: 60 * time.Second,
		})
	})

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")

	engID := "01900000-b002-7000-8000-000000000001"
	scenarioID := "01900000-b002-7000-8000-000000000002"
	stepID := "01900000-b002-7000-8000-000000000003"

	seedBlindEngagementDB(t, server, engID, scenarioID, red, blue)
	seedUnrevealedStep(t, server, stepID, scenarioID)

	ts := httptest.NewServer(server.handler)
	t.Cleanup(ts.Close)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	blueCookie := sessionCookie(t, server.login(blue.Email, testPassword))
	frames := captureSSE(t,
		ts.URL+eventsPathTest+"?topics="+events.EngagementTopic(engID),
		blueCookie, ctx)

	time.Sleep(100 * time.Millisecond)

	// Red sets presence with focus on unrevealed step.
	redCookie := sessionCookie(t, server.login(red.Email, testPassword))
	presenceUUID := uuid.Must(uuid.NewV7())
	stepUUID := uuid.MustParse(stepID)
	setPresence(t, server, engID, presenceUUID, &stepUUID, nil, redCookie)

	// Blue presence.join must have focus stripped.
	fr := waitForFrame(t, frames, events.TypePresenceJoin, 5*time.Second)
	assertPresenceFocusStripped(t, fr, red.ID)

	cancel()
	_ = drainFrames(t, frames, 500*time.Millisecond)
}

// ---------------------------------------------------------------------------
// Test 3: Presence focus NOT stripped in standard engagement
// ---------------------------------------------------------------------------

func TestStandardEngagementPresenceFocusNotStripped(t *testing.T) {
	t.Parallel()
	server := newAuthServerDeps(t, func(d *Deps) {
		d.Presence = presence.New(presence.Options{
			HeartbeatTTL: 60 * time.Second,
		})
	})

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")

	engID := "01900000-b003-7000-8000-000000000001"
	scenarioID := "01900000-b003-7000-8000-000000000002"
	stepID := "01900000-b003-7000-8000-000000000003"

	seedStandardEngagement(t, server, engID, scenarioID, stepID, red, blue)

	redCookie := sessionCookie(t, server.login(red.Email, testPassword))
	presenceUUID := uuid.Must(uuid.NewV7())
	stepUUID := uuid.MustParse(stepID)
	setPresence(t, server, engID, presenceUUID, &stepUUID, nil, redCookie)

	// Blue sees focus intact in standard engagement.
	blueCookie := sessionCookie(t, server.login(blue.Email, testPassword))
	assertFocusStripped(t, server, engID, blueCookie, true)
}

// ---------------------------------------------------------------------------
// Test 4: Hub Allow filter drops unrevealed step events for blue
// ---------------------------------------------------------------------------

func TestBlindHubAllowFilterDropsUnrevealedActivity(t *testing.T) {
	t.Parallel()

	hub := events.NewHub(events.Options{})
	blueScope := blind.Scope{Blind: true, Seat: authz.EngagementRoleBlue}

	ch, unsub, err := hub.Subscribe(t.Context(), events.Subscription{
		Topics: []string{events.EngagementTopic("01900000-b004-7000-8000-000000000001")},
		Allow: func(ev events.Event) bool {
			return events.VisibleActivity(blueScope, ev)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer unsub()

	revealedTrue := true
	revealedFalse := false

	// Publish an unrevealed step event — should be dropped.
	unrevealedData, err := json.Marshal(events.EventData{
		ObjectType: events.ObjectStep,
		ObjectID:   "step-1",
		Revealed:   &revealedFalse,
	})
	if err != nil {
		t.Fatal(err)
	}
	hub.Publish(events.EngagementTopic("01900000-b004-7000-8000-000000000001"), events.Event{
		Type: "step.created",
		Data: unrevealedData,
	})

	// Publish a revealed step event — should arrive.
	revealedData, err := json.Marshal(events.EventData{
		ObjectType: events.ObjectStep,
		ObjectID:   "step-2",
		Revealed:   &revealedTrue,
	})
	if err != nil {
		t.Fatal(err)
	}
	hub.Publish(events.EngagementTopic("01900000-b004-7000-8000-000000000001"), events.Event{
		Type: "step.revealed",
		Data: revealedData,
	})

	// Only the revealed event should arrive.
	select {
	case ev := <-ch:
		var d events.EventData
		if err := json.Unmarshal(ev.Data, &d); err != nil {
			t.Fatal(err)
		}
		if d.ObjectID != "step-2" {
			t.Errorf("got event for step %s, want step-2 (unrevealed step-1 should be filtered)", d.ObjectID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for revealed step event")
	}

	// Verify no second event (the unrevealed one was dropped).
	select {
	case ev := <-ch:
		var d events.EventData
		if err := json.Unmarshal(ev.Data, &d); err != nil {
			t.Errorf("unexpected second event (unmarshal failed: %v)", err)
		} else {
			t.Errorf("unexpected second event for step %s", d.ObjectID)
		}
	case <-time.After(200 * time.Millisecond):
		// Good — unrevealed event was dropped.
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func createUser(t *testing.T, server *authServer, email, displayName string) identity.User {
	t.Helper()
	user, err := identity.NewUsers(server.db).Create(t.Context(), identity.NewUser{
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: testPasswordHash(),
		PlatformRole: authz.PlatformRoleMember,
		Status:       identity.StatusActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	return user
}

func seedUnrevealedStep(t *testing.T, server *authServer, stepID, scenarioID string) {
	t.Helper()
	if err := server.db.Write(t.Context(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.step (id, scenario_id, ordinal, name, objective, technique_id,
			 subtechnique_id, tactic_id, "procedure", template_id, target_asset,
			 tools, controls_in_scope, attack_version, revealed_at, created_at, updated_at)
			VALUES (?, ?, 0, 'hidden-step', '', 'T1003', '', '',
			 '{}', '', '', '[]', '[]', '15.1',
			 NULL, '2026-01-01 00:00:00', '2026-01-01 00:00:00')`,
			stepID, scenarioID,
		)
		return err
	}); err != nil {
		t.Fatalf("seeding step: %v", err)
	}
}

func seedRevealedStep(t *testing.T, server *authServer, stepID, scenarioID string) {
	t.Helper()
	if err := server.db.Write(t.Context(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.step (id, scenario_id, ordinal, name, objective, technique_id,
			 subtechnique_id, tactic_id, "procedure", template_id, target_asset,
			 tools, controls_in_scope, attack_version, revealed_at, created_at, updated_at)
			VALUES (?, ?, 1, 'revealed-step', '', 'T1003', '', '',
			 '{}', '', '', '[]', '[]', '15.1',
			 '2026-01-02 00:00:00', '2026-01-01 00:00:00', '2026-01-01 00:00:00')`,
			stepID, scenarioID,
		)
		return err
	}); err != nil {
		t.Fatalf("seeding revealed step: %v", err)
	}
}

func seedStandardEngagement(t *testing.T, server *authServer,
	engID, scenarioID, stepID string, red, blue identity.User,
) {
	t.Helper()
	if err := server.db.Write(t.Context(), func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.engagement (id, name, client, description, status,
			 starts_on, ends_on, attack_version, mode, auto_reveal_on_start,
			 created_by, created_at, updated_at)
			VALUES (?, 'standard-test', 'test', '', 'active',
			 '2026-01-01', '2026-06-01', '15.1', 'standard', false,
			 ?, '2026-01-01 00:00:00', '2026-01-01 00:00:00')`,
			engID, red.ID,
		)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(t.Context(),
			`INSERT INTO app.scenario (id, engagement_id, ordinal, name, narrative,
			 source, threat_actor, source_ref, plan_id, created_at, updated_at)
			VALUES (?, ?, 0, 'scenario-1', '',
			 'manual', '', '', '', '2026-01-01 00:00:00', '2026-01-01 00:00:00')`,
			scenarioID, engID,
		)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(t.Context(),
			`INSERT INTO app.step (id, scenario_id, ordinal, name, objective, technique_id,
			 subtechnique_id, tactic_id, "procedure", template_id, target_asset,
			 tools, controls_in_scope, attack_version, revealed_at, created_at, updated_at)
			VALUES (?, ?, 0, 'visible-step', '', 'T1003', '', '',
			 '{}', '', '', '[]', '[]', '15.1',
			 NULL, '2026-01-01 00:00:00', '2026-01-01 00:00:00')`,
			stepID, scenarioID,
		)
		return err
	}); err != nil {
		t.Fatal(err)
	}

	memberships := identity.NewMemberships(server.db)
	if _, err := memberships.Add(t.Context(), identity.NewMembership{
		EngagementID: engID, UserID: red.ID, Role: authz.EngagementRoleRed,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := memberships.Add(t.Context(), identity.NewMembership{
		EngagementID: engID, UserID: blue.ID, Role: authz.EngagementRoleBlue,
	}); err != nil {
		t.Fatal(err)
	}
}

func setPresence(t *testing.T, server *authServer, engID string,
	presenceUUID uuid.UUID, stepID, execID *uuid.UUID, cookie *http.Cookie,
) {
	t.Helper()
	payload := map[string]any{
		"presenceId": presenceUUID.String(),
	}
	if stepID != nil || execID != nil {
		focus := map[string]any{}
		if stepID != nil {
			focus["stepId"] = stepID.String()
		}
		if execID != nil {
			focus["executionId"] = execID.String()
		}
		payload["focus"] = focus
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	res := server.send(http.MethodPut,
		BasePath+"/engagements/"+engID+"/presence?presenceId="+presenceUUID.String(),
		string(body), cookie)
	if res.Code != http.StatusNoContent {
		t.Fatalf("set presence = %d\n%s", res.Code, res.Body)
	}
}

func assertFocusStripped(t *testing.T, server *authServer, engID string,
	cookie *http.Cookie, wantFocus bool,
) {
	t.Helper()
	res := server.get(BasePath+"/engagements/"+engID+"/presence", cookie)
	if res.Code != http.StatusOK {
		t.Fatalf("GET presence = %d\n%s", res.Code, res.Body)
	}
	var snapshot struct {
		Entries []gen.PresenceEntry `json:"entries"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Entries) == 0 {
		t.Fatal("presence snapshot is empty")
	}
	hasFocus := snapshot.Entries[0].Focus != nil && snapshot.Entries[0].Focus.StepId != nil
	if hasFocus != wantFocus {
		t.Errorf("focus present = %v, want %v", hasFocus, wantFocus)
	}
}

func assertPresenceFocusStripped(t *testing.T, fr sseFrame, expectedUserID string) {
	t.Helper()
	env := parseEnvelope(t, fr)
	var pd struct {
		StepID      string `json:"stepId"`
		ExecutionID string `json:"executionId"`
		UserID      string `json:"userId"`
	}
	if err := json.Unmarshal(env.Data, &pd); err != nil {
		t.Fatalf("unmarshal presence data: %v", err)
	}
	if pd.UserID != expectedUserID {
		t.Errorf("presence userId = %s, want %s", pd.UserID, expectedUserID)
	}
	if pd.StepID != "" {
		t.Errorf("blue saw stepId=%s in presence SSE, want empty (stripped for blind)", pd.StepID)
	}
}
