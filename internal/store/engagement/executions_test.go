package engagement_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/store/engagement"
)

func TestPatchRedLifecycle(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	eng := mustCreateEngagement(t, r)
	sc := mustCreateScenario(t, r, eng.ID, 1, "Scenario")
	_, exec := mustCreateStepWithExecution(t, r, sc.ID, 1, "Step")

	// An empty patch still advances the optimistic-lock version.
	patched, err := r.Executions.PatchRed(ctx, exec.ID, exec.Version, engagement.RedPatchChanges{})
	if err != nil {
		t.Fatal(err)
	}
	if patched.Version != exec.Version+1 {
		t.Fatalf("empty patch version = %d, want %d", patched.Version, exec.Version+1)
	}

	started := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	ended := started.Add(45 * time.Minute)
	status := engagement.ExecutionStatusRunning
	command, source, target := "whoami", "kali-1", "dc-1"
	notes, executed := "clean sweep", "red-1"
	patched, err = r.Executions.PatchRed(ctx, exec.ID, patched.Version, engagement.RedPatchChanges{
		Status: &status, StartedAt: &started, EndedAt: &ended,
		CommandRun: &command, SourceHost: &source, TargetHost: &target,
		RedNotes: &notes, ExecutedBy: &executed,
	})
	if err != nil {
		t.Fatal(err)
	}
	if patched.Status != status || patched.StartedAt == nil || !patched.StartedAt.Equal(started) || patched.EndedAt == nil || !patched.EndedAt.Equal(ended) {
		t.Fatalf("red timing fields did not persist: %+v", patched)
	}
	if patched.CommandRun != command || patched.SourceHost != source || patched.TargetHost != target || patched.RedNotes != notes || patched.ExecutedBy != executed {
		t.Fatalf("red text fields did not persist: %+v", patched)
	}

	// A second patch that omits every field must preserve them.
	patched, err = r.Executions.PatchRed(ctx, exec.ID, patched.Version, engagement.RedPatchChanges{})
	if err != nil {
		t.Fatal(err)
	}
	got, err := r.Executions.ByID(ctx, exec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.CommandRun != command || got.RedNotes != notes || got.ExecutedBy != executed || got.StartedAt == nil || !got.StartedAt.Equal(started) || got.EndedAt == nil || !got.EndedAt.Equal(ended) {
		t.Fatalf("omitted fields were overwritten: %+v", got)
	}
	if got.Version != patched.Version || got.Status != status {
		t.Fatalf("version/status after empty patch: %+v", got)
	}
}

func TestPatchBlueLifecycle(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	eng := mustCreateEngagement(t, r)
	sc := mustCreateScenario(t, r, eng.ID, 1, "Scenario")
	_, exec := mustCreateStepWithExecution(t, r, sc.ID, 1, "Step")

	detected := time.Date(2026, 5, 1, 10, 30, 0, 0, time.UTC)
	scored := detected.Add(2 * time.Hour)
	category := engagement.DetectionCategoryTechnique
	protection := engagement.ProtectionBlocked
	sev, rule, source := "high", "rule-42", "elastic"
	notes, scoredBy := "detected on first alert", "blue-1"
	modifiers := json.RawMessage(`["alerted","logged"]`)
	patched, err := r.Executions.PatchBlue(ctx, exec.ID, exec.Version, engagement.BluePatchChanges{
		DetectionCategory: &category, DetectionModifiers: modifiers, Protection: &protection,
		DetectedAt: &detected, DetectingSource: &source, DetectingRuleRef: &rule,
		AlertSeverity: &sev, BlueNotes: &notes, ScoredBy: &scoredBy, ScoredAt: &scored,
	})
	if err != nil {
		t.Fatal(err)
	}
	if patched.DetectionCategory == nil || *patched.DetectionCategory != category || patched.Protection == nil || *patched.Protection != protection || patched.DetectedAt == nil || !patched.DetectedAt.Equal(detected) || patched.ScoredAt == nil || !patched.ScoredAt.Equal(scored) {
		t.Fatalf("typed blue fields did not persist: %+v", patched)
	}
	if patched.DetectingSource != source || patched.DetectingRuleRef != rule || patched.AlertSeverity != sev || patched.BlueNotes != notes || patched.ScoredBy != scoredBy {
		t.Fatalf("text blue fields did not persist: %+v", patched)
	}
	if string(patched.DetectionModifiers) != string(modifiers) {
		t.Fatalf("modifiers = %s, want %s", patched.DetectionModifiers, modifiers)
	}

	// An explicit empty array is a value, not "leave unchanged".
	empty := json.RawMessage(`[]`)
	patched, err = r.Executions.PatchBlue(ctx, exec.ID, patched.Version, engagement.BluePatchChanges{DetectionModifiers: empty})
	if err != nil {
		t.Fatal(err)
	}
	if string(patched.DetectionModifiers) != `[]` {
		t.Fatalf("empty modifiers = %s, want []", patched.DetectionModifiers)
	}
	// Fields omitted by the second patch keep their earlier values.
	got, err := r.Executions.ByStepID(ctx, exec.StepID)
	if err != nil {
		t.Fatal(err)
	}
	if got.DetectionCategory == nil || *got.DetectionCategory != category || got.BlueNotes != notes || got.ScoredBy != scoredBy {
		t.Fatalf("unrelated blue fields were overwritten: %+v", got)
	}
}

func TestPatchVersionConflicts(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	eng := mustCreateEngagement(t, r)
	sc := mustCreateScenario(t, r, eng.ID, 1, "Scenario")
	_, exec := mustCreateStepWithExecution(t, r, sc.ID, 1, "Step")
	// Move the row to version 2 with one empty patch.
	if _, err := r.Executions.IncrementVersion(ctx, exec.ID, exec.Version); err != nil {
		t.Fatal(err)
	}
	stale := exec.Version
	status := engagement.ExecutionStatusRunning
	notes := "conflicting write"
	for name, patch := range map[string]func() error{
		"red": func() error {
			_, err := r.Executions.PatchRed(ctx, exec.ID, stale, engagement.RedPatchChanges{Status: &status})
			return err
		},
		"blue": func() error {
			_, err := r.Executions.PatchBlue(ctx, exec.ID, stale, engagement.BluePatchChanges{BlueNotes: &notes})
			return err
		},
		"bump": func() error { _, err := r.Executions.IncrementVersion(ctx, exec.ID, stale); return err },
	} {
		t.Run(name, func(t *testing.T) {
			err := patch()
			if err == nil || !strings.Contains(err.Error(), "version conflict") {
				t.Fatalf("stale version accepted: %v", err)
			}
		})
	}
	// No losing write landed.
	got, err := r.Executions.ByID(ctx, exec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != exec.Version+1 || got.Status != engagement.ExecutionStatusPending {
		t.Fatalf("losing write landed: %+v", got)
	}
}

func TestListByEngagementFiltersAndOrder(t *testing.T) {
	r := newRepos(t)
	ctx := context.Background()
	eng := mustCreateEngagement(t, r)
	other := mustCreateEngagement(t, r)
	scA := mustCreateScenario(t, r, eng.ID, 1, "Scenario A")
	scB := mustCreateScenario(t, r, eng.ID, 2, "Scenario B")
	otherSc := mustCreateScenario(t, r, other.ID, 1, "Other")
	// Deliberately create workbook-later steps first: order must come from
	// ordinals, not insertion order.
	_, laterInA := mustCreateStepWithExecution(t, r, scA.ID, 2, "A2")
	_, firstInA := mustCreateStepWithExecution(t, r, scA.ID, 1, "A1")
	_, onlyInB := mustCreateStepWithExecution(t, r, scB.ID, 1, "B1")
	_, otherExec := mustCreateStepWithExecution(t, r, otherSc.ID, 1, "Other")
	complete := engagement.ExecutionStatusComplete
	if _, err := r.Executions.PatchRed(ctx, laterInA.ID, laterInA.Version, engagement.RedPatchChanges{Status: &complete}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name       string
		scenarioID *string
		status     *engagement.ExecutionStatus
		want       []string
	}{
		{"workbook order", nil, nil, []string{firstInA.ID, laterInA.ID, onlyInB.ID}},
		{"scenario filter", &scA.ID, nil, []string{firstInA.ID, laterInA.ID}},
		{"status filter", nil, &complete, []string{laterInA.ID}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			list, err := r.Executions.ListByEngagement(ctx, eng.ID, tc.scenarioID, tc.status)
			if err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(list))
			for _, e := range list {
				got = append(got, e.ID)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("list = %v, want %v", got, tc.want)
			}
		})
	}
	// Executions never leak across engagements.
	list, err := r.Executions.ListByEngagement(ctx, other.ID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != otherExec.ID {
		t.Fatalf("cross-engagement leak: %+v", list)
	}
	// An unknown execution reports not found.
	if _, err := r.Executions.ByID(ctx, "01900000-0000-7000-e000-0000000000aa"); !errors.Is(err, apierr.ErrNotFound) {
		t.Fatalf("unknown execution: %v", err)
	}
}
