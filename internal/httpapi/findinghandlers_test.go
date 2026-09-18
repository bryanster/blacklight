package httpapi

import (
	"database/sql"
	"slices"
	"testing"

	"github.com/bryanster/blacklight/internal/httpapi/gen"
)

// TestFindingStepIdsNotFilteredOutsideBlindBlue pins the other half of the
// BL-007 step-id filter: only the blue seat of a blind engagement is withheld
// unrevealed links, so a standard engagement's finding arrives with its
// complete link set — even when the linked step is unrevealed, which a
// standard engagement simply does not hide.
func TestFindingStepIdsNotFilteredOutsideBlindBlue(t *testing.T) {
	server := newAuthServer(t)

	red := createUser(t, server, "red@example.com", "RedUser")
	blue := createUser(t, server, "blue@example.com", "BlueUser")

	const (
		engID      = "01900000-b001-7000-8000-000000000011"
		scenarioID = "01900000-b001-7000-8000-000000000012"
		stepID     = "01900000-b001-7000-8000-000000000013"
		findingID  = "01900000-b001-7000-8000-000000000014"
		noLinksID  = "01900000-b001-7000-8000-000000000015"
	)

	seedStandardEngagement(t, server, engID, scenarioID, stepID, red, blue)

	if err := server.db.Write(t.Context(), func(tx *sql.Tx) error {
		for _, id := range []string{findingID, noLinksID} {
			if _, err := tx.ExecContext(t.Context(),
				`INSERT INTO app.finding (id, engagement_id, title, description, severity,
				 recommendation, "owner", status, created_from_execution, created_at, updated_at)
				 VALUES (?, ?, 'Plain Finding', '', 'high', '', ?, 'open', NULL, NOW(), NOW())`,
				id, engID, red.ID); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(t.Context(),
			`INSERT INTO app.finding_step (finding_id, step_id) VALUES (?, ?)`,
			findingID, stepID)
		return err
	}); err != nil {
		t.Fatalf("seeding findings: %v", err)
	}

	blueCookie := sessionCookie(t, server.login(blue.Email, testPassword))

	list := decodeJSON[[]gen.Finding](t, server.get(BasePath+"/engagements/"+engID+"/findings", blueCookie))
	if len(list) != 2 {
		t.Fatalf("blue's findings list has %d entries, want 2", len(list))
	}
	for _, f := range list {
		ids := stepIDStrings(f.StepIds)
		want := f.Id.String() == findingID
		if slices.Contains(ids, stepID) != want {
			t.Errorf("finding %s stepIds = %v, linked = %v — filter misapplied", f.Id, ids, want)
		}
	}

	got := decodeJSON[gen.Finding](t, server.get(BasePath+"/findings/"+findingID, blueCookie))
	if ids := stepIDStrings(got.StepIds); !slices.Contains(ids, stepID) {
		t.Errorf("GET finding stepIds = %v, want it to contain %s — filter over-applied", ids, stepID)
	}
}
