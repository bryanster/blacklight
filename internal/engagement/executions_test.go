package engagement_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/authn"
	"github.com/bryanster/blacklight/internal/engagement"
	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	storengagement "github.com/bryanster/blacklight/internal/store/engagement"
	"github.com/bryanster/blacklight/internal/store/identity"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

type execFixture struct {
	svc  *engagement.Service
	eng  storengagement.Engagement
	step storengagement.Step
	exec storengagement.Execution
}

func newExecFixture(t *testing.T, mode storengagement.EngagementMode, autoReveal bool) execFixture {
	t.Helper()
	db := storetest.Migrated(t)
	engagements := storengagement.NewEngagements(db)
	memberships := identity.NewMemberships(db)
	users := identity.NewUsers(db)
	scenarios := storengagement.NewScenarios(db)
	steps := storengagement.NewSteps(db)
	executions := storengagement.NewExecutions(db)
	svc, err := engagement.New(engagement.Deps{
		Engagements: engagements,
		Memberships: memberships,
		Scenarios:   scenarios,
		Steps:       steps,
		Executions:  executions,
		Users:       users,
	})
	if err != nil {
		t.Fatalf("engagement.New: %v", err)
	}
	eng, err := engagements.Create(context.Background(), storengagement.NewEngagement{
		Name: "Exec service", AttackVersion: "15.1", Mode: mode,
		AutoRevealOnStart: autoReveal, CreatedBy: "user-1",
	})
	if err != nil {
		t.Fatalf("create engagement: %v", err)
	}
	sc, err := scenarios.Create(context.Background(), storengagement.NewScenario{
		EngagementID: eng.ID, Ordinal: 1, Name: "Scenario",
		Narrative: "n", Source: storengagement.ScenarioSourceManual, ThreatActor: "APT999",
	})
	if err != nil {
		t.Fatalf("create scenario: %v", err)
	}
	step, exec, err := steps.CreateWithExecution(context.Background(), storengagement.NewStep{
		ScenarioID: sc.ID, Ordinal: 1, Name: "Step", Objective: "obj",
		TechniqueID: "T1059", SubtechniqueID: "T1059.001", TacticID: "TA0002",
		Procedure:   json.RawMessage(`{"executor":"sh","command":"whoami"}`),
		TargetAsset: "host-1", Tools: json.RawMessage(`["nmap"]`),
		ControlsInScope: json.RawMessage(`["EDR"]`), AttackVersion: "15.1",
	})
	if err != nil {
		t.Fatalf("create step: %v", err)
	}
	return execFixture{svc: svc, eng: eng, step: step, exec: exec}
}

func statusPtr(s storengagement.ExecutionStatus) *storengagement.ExecutionStatus { return &s }

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestPatchRedExecutionLifecycle(t *testing.T) {
	f := newExecFixture(t, storengagement.EngagementModeBlind, true)
	ctx := context.Background()
	actor := authn.Subject{UserID: "red-1"}

	// pending → running with no explicit time: server sets started_at and
	// executed_by, and the blind engagement auto-reveals the step.
	got, err := f.svc.PatchRedExecution(ctx, actor, f.eng.ID, f.exec.ID, engagement.PatchRedExecutionInput{
		Version: f.exec.Version, Status: statusPtr(storengagement.ExecutionStatusRunning),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != storengagement.ExecutionStatusRunning || got.StartedAt == nil || got.ExecutedBy != "red-1" {
		t.Fatalf("transition fields: %+v", got)
	}
	step, err := f.svc.GetStep(ctx, f.step.ID)
	if err != nil {
		t.Fatal(err)
	}
	if step.RevealedAt == nil {
		t.Fatal("auto-reveal did not fire")
	}

	// running → complete with an explicit end time and a note.
	ended := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	got, err = f.svc.PatchRedExecution(ctx, actor, f.eng.ID, f.exec.ID, engagement.PatchRedExecutionInput{
		Version: got.Version, Status: statusPtr(storengagement.ExecutionStatusComplete),
		EndedAt: &ended, RedNotes: new("done"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != storengagement.ExecutionStatusComplete || got.EndedAt == nil || !got.EndedAt.Equal(ended) || got.RedNotes != "done" {
		t.Fatalf("completion fields: %+v", got)
	}

	// Terminal state: note edits allowed, status changes refused.
	got, err = f.svc.PatchRedExecution(ctx, actor, f.eng.ID, f.exec.ID, engagement.PatchRedExecutionInput{
		Version: got.Version, RedNotes: new("note only"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.RedNotes != "note only" {
		t.Fatalf("terminal note edit lost: %+v", got)
	}
	_, err = f.svc.PatchRedExecution(ctx, actor, f.eng.ID, f.exec.ID, engagement.PatchRedExecutionInput{
		Version: got.Version, Status: statusPtr(storengagement.ExecutionStatusRunning),
	})
	if !strings.Contains(errString(err), "terminal state") {
		t.Fatalf("terminal transition accepted: %v", err)
	}

	// Stale version loses.
	_, err = f.svc.PatchRedExecution(ctx, actor, f.eng.ID, f.exec.ID, engagement.PatchRedExecutionInput{
		Version: 1, RedNotes: new("stale"),
	})
	if !strings.Contains(errString(err), "version conflict") {
		t.Fatalf("stale version accepted: %v", err)
	}

	// Wrong path engagement is indistinguishable from a missing id.
	_, err = f.svc.PatchRedExecution(ctx, actor, "01900000-0000-7000-e000-0000000000aa", f.exec.ID, engagement.PatchRedExecutionInput{Version: got.Version})
	var typed *apierr.Error
	if !errors.As(err, &typed) || typed.Status() != 404 {
		t.Fatalf("cross-engagement patch: %v", err)
	}

	// Closed engagements refuse writes.
	if _, err := f.svc.SetStatus(ctx, actor, f.eng.ID, storengagement.EngagementStatusClosed); err != nil {
		t.Fatal(err)
	}
	_, err = f.svc.PatchRedExecution(ctx, actor, f.eng.ID, f.exec.ID, engagement.PatchRedExecutionInput{Version: got.Version})
	if !strings.Contains(errString(err), "engagement is closed") {
		t.Fatalf("closed engagement patch accepted: %v", err)
	}
}

func TestPatchBlueDetectionLifecycle(t *testing.T) {
	f := newExecFixture(t, storengagement.EngagementModeStandard, false)
	ctx := context.Background()
	redActor := authn.Subject{UserID: "red-1"}
	blue := authn.Subject{UserID: "blue-1"}
	started := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	patched, err := f.svc.PatchRedExecution(ctx, redActor, f.eng.ID, f.exec.ID, engagement.PatchRedExecutionInput{
		Version: f.exec.Version, Status: statusPtr(storengagement.ExecutionStatusRunning), StartedAt: &started,
	})
	if err != nil {
		t.Fatal(err)
	}

	category := storengagement.DetectionCategoryTechnique
	protection := storengagement.ProtectionBlocked
	sev := "high"
	detected := started.Add(30 * time.Minute)
	modifiers := []string{"correlated", "correlated", "alert"}
	got, err := f.svc.PatchBlueDetection(ctx, blue, f.eng.ID, f.exec.ID, engagement.PatchBlueDetectionInput{
		Version: patched.Version, DetectionCategory: &category, Protection: &protection,
		AlertSeverity: &sev, DetectionModifiers: &modifiers, DetectedAt: &detected,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.DetectionCategory == nil || *got.DetectionCategory != category || got.Protection == nil || *got.Protection != protection || got.AlertSeverity != sev {
		t.Fatalf("detection fields: %+v", got)
	}
	if got.ScoredBy != blue.UserID || got.ScoredAt == nil {
		t.Fatalf("scoring stamp not set: %+v", got)
	}
	var mods []string
	if err := json.Unmarshal(got.DetectionModifiers, &mods); err != nil {
		t.Fatal(err)
	}
	if len(mods) != 2 || mods[0] != "correlated" || mods[1] != "alert" {
		t.Fatalf("modifiers not deduped in order: %s", got.DetectionModifiers)
	}

	// detected_at before started_at → validation error.
	early := started.Add(-time.Minute)
	_, err = f.svc.PatchBlueDetection(ctx, blue, f.eng.ID, f.exec.ID, engagement.PatchBlueDetectionInput{
		Version: got.Version, DetectedAt: &early,
	})
	var typed *apierr.Error
	if !errors.As(err, &typed) || typed.Status() != 400 {
		t.Fatalf("early detectedAt accepted: %v", err)
	}

	// Clearing modifiers with an explicit empty array is valid.
	got, err = f.svc.PatchBlueDetection(ctx, blue, f.eng.ID, f.exec.ID, engagement.PatchBlueDetectionInput{
		Version: got.Version, DetectionModifiers: &[]string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(got.DetectionModifiers) != `[]` {
		t.Fatalf("empty modifiers = %s", got.DetectionModifiers)
	}

	// Cross-engagement path is a 404.
	_, err = f.svc.PatchBlueDetection(ctx, blue, "01900000-0000-7000-e000-0000000000aa", f.exec.ID, engagement.PatchBlueDetectionInput{Version: got.Version})
	if !errors.As(err, &typed) || typed.Status() != 404 {
		t.Fatalf("cross-engagement blue patch: %v", err)
	}

	// GetExecutionInEngagement binds the path engagement.
	if _, err := f.svc.GetExecutionInEngagement(ctx, f.eng.ID, f.exec.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.GetExecutionInEngagement(ctx, "01900000-0000-7000-e000-0000000000aa", f.exec.ID); !errors.Is(err, apierr.ErrNotFound) {
		t.Fatalf("wrong-engagement get: %v", err)
	}
}
