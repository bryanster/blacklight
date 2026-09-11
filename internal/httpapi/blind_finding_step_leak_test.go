package httpapi

import (
	"database/sql"
	"net/http"
	"slices"
	"testing"

	"github.com/oapi-codegen/runtime/types"

	"github.com/bryanster/blacklight/internal/httpapi/gen"
)

// TestBlindFindingStepIdsLeakUnrevealedSteps pins the BL-007 regression: in a
// blind engagement the findings endpoints must not hand a blue member the ids
// of unrevealed steps.
//
// It began as the demonstration of exactly that leak: finding.read carries no
// GuardBlindMode and findingToWire returned every linked step id, so blue's
// findings list disclosed ids the step endpoint (correctly) answers 404 on.
// findingToWire now filters live — unrevealed links are withheld, revealed
// ones survive, and red still sees everything.
func TestBlindFindingStepIdsLeakUnrevealedSteps(t *testing.T) {
	server := newAuthServer(t)

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")

	const (
		engID      = "01900000-b001-7000-8000-000000000001"
		scenarioID = "01900000-b001-7000-8000-000000000002"
		hiddenID   = "01900000-b001-7000-8000-000000000003"
		findingID  = "01900000-b001-7000-8000-000000000004"
		visibleID  = "01900000-b001-7000-8000-000000000005"
	)

	seedBlindEngagementDB(t, server, engID, scenarioID, red, blue)
	seedUnrevealedStep(t, server, hiddenID, scenarioID)
	seedRevealedStep(t, server, visibleID, scenarioID)

	// Link one finding to both the unrevealed and the revealed step.
	if err := server.db.Write(t.Context(), func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.finding (id, engagement_id, title, description, severity,
			 recommendation, "owner", status, created_from_execution, created_at, updated_at)
			 VALUES (?, ?, 'Hidden Finding', '', 'high', '', ?, 'open', NULL, NOW(), NOW())`,
			findingID, engID, red.ID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.finding_step (finding_id, step_id) VALUES (?, ?)`,
			findingID, hiddenID); err != nil {
			return err
		}
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.finding_step (finding_id, step_id) VALUES (?, ?)`,
			findingID, visibleID)
		return err
	}); err != nil {
		t.Fatalf("seeding finding: %v", err)
	}

	blueCookie := sessionCookie(t, server.login(blue.Email, testPassword))

	// Blind mode conceals the unrevealed step itself and shows the revealed one.
	if rec := server.get(BasePath+"/engagements/"+engID+"/scenarios/"+scenarioID+"/steps/"+hiddenID, blueCookie); rec.Code != http.StatusNotFound {
		t.Fatalf("blue fetching the unrevealed step = %d, want 404 (blind conceal)\nbody: %s",
			rec.Code, rec.Body)
	}
	if rec := server.get(BasePath+"/engagements/"+engID+"/scenarios/"+scenarioID+"/steps/"+visibleID, blueCookie); rec.Code != http.StatusOK {
		t.Fatalf("blue fetching the revealed step = %d, want 200\nbody: %s",
			rec.Code, rec.Body)
	}

	// The findings list keeps the finding but withholds only the hidden link.
	list := decodeJSON[[]gen.Finding](t, server.get(BasePath+"/engagements/"+engID+"/findings", blueCookie))
	if len(list) != 1 {
		t.Fatalf("blue's findings list has %d entries, want 1", len(list))
	}
	assertBlueFindingStepIDs(t, list[0].StepIds, visibleID, hiddenID)

	// The single-finding read is filtered the same way.
	got := decodeJSON[gen.Finding](t, server.get(BasePath+"/findings/"+findingID, blueCookie))
	assertBlueFindingStepIDs(t, got.StepIds, visibleID, hiddenID)

	// (The patch response renders through the same findingToWire; it is not
	// driven here because the store's finding update currently cannot rewrite
	// a step-linked finding at all — an unrelated DuckDB foreign-key failure
	// for every seat.)

	// Red is not held to blind mode and keeps every link.
	redCookie := sessionCookie(t, server.login(red.Email, testPassword))
	redList := decodeJSON[[]gen.Finding](t, server.get(BasePath+"/engagements/"+engID+"/findings", redCookie))
	if len(redList) != 1 {
		t.Fatalf("red's findings list has %d entries, want 1", len(redList))
	}
	redIDs := stepIDStrings(redList[0].StepIds)
	for _, want := range []string{hiddenID, visibleID} {
		if !slices.Contains(redIDs, want) {
			t.Errorf("red's finding stepIds missing %s — blind filter over-applied: %v", want, redIDs)
		}
	}
}

// assertBlueFindingStepIDs checks that a blue-seat finding carries the
// revealed step link and none of the unrevealed ones.
func assertBlueFindingStepIDs(t *testing.T, got []types.UUID, revealedID, hiddenID string) {
	t.Helper()
	ids := stepIDStrings(got)
	if slices.Contains(ids, hiddenID) {
		t.Errorf("blue's finding stepIds contains unrevealed step %s — leak reproduced: %v", hiddenID, ids)
	}
	if !slices.Contains(ids, revealedID) {
		t.Errorf("blue's finding stepIds missing revealed step %s — filter over-applied: %v", revealedID, ids)
	}
}

func stepIDStrings(ids []types.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}
