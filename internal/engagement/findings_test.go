package engagement_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bryanster/blacklight/internal/engagement"
	storengagement "github.com/bryanster/blacklight/internal/store/engagement"
)

func TestCreateFindingRequiresATitleAndAnEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	if _, err := d.svc.CreateFinding(ctx, d.actor(), engagement.CreateFindingInput{
		EngagementID: d.eng.ID,
	}); err == nil {
		t.Fatal("a finding with no title was accepted")
	}

	if _, err := d.svc.CreateFinding(ctx, d.actor(), engagement.CreateFindingInput{
		Title: "detection gap",
	}); err == nil {
		t.Fatal("a finding with no engagement was accepted")
	}
}

func TestCreateFindingDefaultsOwnerSeverityAndStatus(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	f, err := d.svc.CreateFinding(context.Background(), d.actor(), engagement.CreateFindingInput{
		EngagementID: d.eng.ID,
		Title:        "missed T1190",
	})
	if err != nil {
		t.Fatalf("create finding: %v", err)
	}
	if f.Owner != workbookActorID {
		t.Errorf("owner = %q, want the raising actor %q", f.Owner, workbookActorID)
	}
	if f.Severity != "medium" {
		t.Errorf("severity = %q, want the medium default", f.Severity)
	}
	if f.Status != storengagement.FindingStatusOpen {
		t.Errorf("status = %q, want %q", f.Status, storengagement.FindingStatusOpen)
	}
	if f.EngagementID != d.eng.ID {
		t.Errorf("engagement = %q, want %q", f.EngagementID, d.eng.ID)
	}
}

func TestCreateFindingRefusesClosedEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	d.closeEngagement(t, d.eng.ID)

	_, err := d.svc.CreateFinding(context.Background(), d.actor(), engagement.CreateFindingInput{
		EngagementID: d.eng.ID,
		Title:        "too late",
	})
	if err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("err = %v, want a closed-engagement refusal", err)
	}
}

func TestUpdateFindingPatchesOnlyTheProvidedFields(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	f := d.createFinding(t, "original")

	got, err := d.svc.UpdateFinding(ctx, d.actor(), f.ID, engagement.UpdateFindingInput{
		Title:  "renamed",
		Status: string(storengagement.FindingStatusResolved),
	})
	if err != nil {
		t.Fatalf("update finding: %v", err)
	}
	if got.Title != "renamed" {
		t.Errorf("title = %q, want renamed", got.Title)
	}
	if got.Status != storengagement.FindingStatusResolved {
		t.Errorf("status = %q, want resolved", got.Status)
	}
	// Fields left out of the patch keep their previous value.
	if got.Severity != "medium" || got.Owner != workbookActorID {
		t.Errorf("unset fields changed: severity=%q owner=%q", got.Severity, got.Owner)
	}
}

func TestUpdateFindingOnAClosedEngagementAllowsOnlyStatus(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	f := d.createFinding(t, "original")
	d.closeEngagement(t, d.eng.ID)

	// Resolving a finding is the one write a closed engagement still accepts.
	got, err := d.svc.UpdateFinding(ctx, d.actor(), f.ID, engagement.UpdateFindingInput{
		Status: string(storengagement.FindingStatusResolved),
	})
	if err != nil {
		t.Fatalf("status change on a closed engagement: %v", err)
	}
	if got.Status != storengagement.FindingStatusResolved {
		t.Fatalf("status = %q, want resolved", got.Status)
	}

	// Anything else is refused.
	_, err = d.svc.UpdateFinding(ctx, d.actor(), f.ID, engagement.UpdateFindingInput{Title: "renamed"})
	if err == nil || !strings.Contains(err.Error(), "only status") {
		t.Fatalf("err = %v, want a status-only refusal", err)
	}
	if current, err := d.svc.GetFinding(ctx, f.ID); err != nil || current.Title != "original" {
		t.Fatalf("refused update mutated the finding: %+v (%v)", current, err)
	}
}

func TestSetFindingStepsLinksStepsAndReportsThem(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	f := d.createFinding(t, "linked")
	extra := d.addStep(t, d.scenario.ID, "step-2", 2)

	if err := d.svc.SetFindingSteps(ctx, d.actor(), f.ID, []string{d.step.ID, extra.ID}); err != nil {
		t.Fatalf("set finding steps: %v", err)
	}

	steps, err := d.svc.FindingSteps(ctx, f.ID)
	if err != nil {
		t.Fatalf("finding steps: %v", err)
	}
	seen := map[string]bool{}
	for _, s := range steps {
		seen[s.ID] = true
	}
	if len(steps) != 2 || !seen[d.step.ID] || !seen[extra.ID] {
		t.Fatalf("finding steps = %+v, want both linked steps", steps)
	}

	// The set replaces rather than appends.
	if err := d.svc.SetFindingSteps(ctx, d.actor(), f.ID, []string{d.step.ID}); err != nil {
		t.Fatalf("replace finding steps: %v", err)
	}
	steps, err = d.svc.FindingSteps(ctx, f.ID)
	if err != nil {
		t.Fatalf("finding steps after replace: %v", err)
	}
	if len(steps) != 1 || steps[0].ID != d.step.ID {
		t.Fatalf("finding steps after replace = %+v, want only step-1", steps)
	}
}

func TestSetFindingStepsRejectsAStepFromAnotherEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	f := d.createFinding(t, "linked")

	foreignScenario := d.addScenario(t, d.other.ID, "foreign", 1)
	foreignStep := d.addStep(t, foreignScenario.ID, "foreign-step", 1)

	err := d.svc.SetFindingSteps(ctx, d.actor(), f.ID, []string{d.step.ID, foreignStep.ID})
	if err == nil || !strings.Contains(err.Error(), "belongs to engagement") {
		t.Fatalf("err = %v, want a cross-engagement refusal", err)
	}
	// The refused call linked nothing, not even the step that was valid.
	steps, err := d.svc.FindingSteps(ctx, f.ID)
	if err != nil {
		t.Fatalf("finding steps: %v", err)
	}
	if len(steps) != 0 {
		t.Fatalf("refused set linked %d steps, want 0", len(steps))
	}
}

func TestSetFindingStepsRejectsAnUnknownStep(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	f := d.createFinding(t, "linked")

	if err := d.svc.SetFindingSteps(context.Background(), d.actor(), f.ID, []string{unknownID}); err == nil {
		t.Fatal("an unknown step id was accepted")
	}
}

func TestDeleteFindingRemovesItAndItsStepLinks(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	f := d.createFinding(t, "doomed")
	if err := d.svc.SetFindingSteps(ctx, d.actor(), f.ID, []string{d.step.ID}); err != nil {
		t.Fatalf("set finding steps: %v", err)
	}

	if err := d.svc.DeleteFinding(ctx, d.actor(), f.ID); err != nil {
		t.Fatalf("delete finding: %v", err)
	}
	if _, err := d.svc.GetFinding(ctx, f.ID); err == nil {
		t.Fatal("finding still present after delete")
	} else {
		notFound(t, err)
	}
	// The step itself survives; only the link is gone.
	if _, err := d.svc.GetStep(ctx, d.step.ID); err != nil {
		t.Fatalf("deleting a finding removed its step: %v", err)
	}
	steps, err := d.svc.FindingSteps(ctx, f.ID)
	if err != nil {
		t.Fatalf("finding steps: %v", err)
	}
	if len(steps) != 0 {
		t.Fatalf("step links survived the delete: %+v", steps)
	}
}

func TestListFindingsReturnsTheEngagementsFindings(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	first := d.createFinding(t, "one")
	second := d.createFinding(t, "two")
	// A finding on another engagement is not in this one's list.
	if _, err := d.svc.CreateFinding(ctx, d.actor(), engagement.CreateFindingInput{
		EngagementID: d.other.ID,
		Title:        "elsewhere",
	}); err != nil {
		t.Fatalf("create finding on the other engagement: %v", err)
	}

	got, err := d.svc.ListFindings(ctx, d.eng.ID)
	if err != nil {
		t.Fatalf("list findings: %v", err)
	}
	seen := map[string]bool{}
	for _, f := range got {
		seen[f.ID] = true
	}
	if len(got) != 2 || !seen[first.ID] || !seen[second.ID] {
		t.Fatalf("findings = %+v, want the two on this engagement", got)
	}
}
