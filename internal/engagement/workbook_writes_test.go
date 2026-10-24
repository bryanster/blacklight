package engagement_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bryanster/blacklight/internal/engagement"
	"github.com/bryanster/blacklight/internal/store/blind"
)

// The workbook write guards that sit between a request and the store: the
// path-engagement binding, the closed-engagement freeze, and the soft freeze on
// a step whose execution has left pending.

func TestPatchScenarioUpdatesAndRefusesAWriteToAClosedEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	name := "renamed"
	narrative := "the new story"
	got, err := d.svc.PatchScenario(ctx, d.actor(), d.eng.ID, d.scenario.ID, engagement.PatchScenarioInput{
		Name:      &name,
		Narrative: &narrative,
	})
	if err != nil {
		t.Fatalf("patch scenario: %v", err)
	}
	if got.Name != name || got.Narrative != narrative {
		t.Fatalf("scenario = %+v, want the patched fields", got)
	}

	// The fields left out of the patch are untouched.
	if got.ThreatActor != d.scenario.ThreatActor {
		t.Errorf("threat_actor changed to %q", got.ThreatActor)
	}

	d.closeEngagement(t, d.eng.ID)
	_, err = d.svc.PatchScenario(ctx, d.actor(), d.eng.ID, d.scenario.ID, engagement.PatchScenarioInput{Name: &name})
	conflict(t, err)
}

func TestPatchScenarioCrossEngagementIsNotFound(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	name := "renamed"
	_, err := d.svc.PatchScenario(context.Background(), d.actor(), d.other.ID, d.scenario.ID, engagement.PatchScenarioInput{Name: &name})
	notFound(t, err)
}

func TestPatchStepUpdatesAndRefusesAWriteToAClosedEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	name := "renamed step"
	asset := "host-2"
	tools := json.RawMessage(`["nmap","curl"]`)
	got, err := d.svc.PatchStep(ctx, d.actor(), d.eng.ID, d.step.ID, engagement.PatchStepInput{
		Name:        &name,
		TargetAsset: &asset,
		Tools:       tools,
	})
	if err != nil {
		t.Fatalf("patch step: %v", err)
	}
	if got.Name != name || got.TargetAsset != asset {
		t.Fatalf("step = %+v, want the patched fields", got)
	}
	var gotTools []string
	if err := json.Unmarshal(got.Tools, &gotTools); err != nil {
		t.Fatalf("unmarshal tools: %v", err)
	}
	if len(gotTools) != 2 || gotTools[0] != "nmap" || gotTools[1] != "curl" {
		t.Fatalf("tools = %s, want the patched list", got.Tools)
	}

	d.closeEngagement(t, d.eng.ID)
	_, err = d.svc.PatchStep(ctx, d.actor(), d.eng.ID, d.step.ID, engagement.PatchStepInput{Name: &name})
	conflict(t, err)
}

func TestPatchStepCrossEngagementIsNotFound(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	name := "renamed"
	_, err := d.svc.PatchStep(context.Background(), d.actor(), d.other.ID, d.step.ID, engagement.PatchStepInput{Name: &name})
	notFound(t, err)
}

func TestRevealStepIsIdempotentAndBindsTheEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	got, err := d.svc.RevealStep(ctx, d.actor(), d.eng.ID, d.step.ID)
	if err != nil {
		t.Fatalf("reveal step: %v", err)
	}
	if got.RevealedAt == nil {
		t.Fatal("revealed_at was not set")
	}
	first := *got.RevealedAt

	// Revealing an already-revealed step succeeds and does not move the stamp.
	again, err := d.svc.RevealStep(ctx, d.actor(), d.eng.ID, d.step.ID)
	if err != nil {
		t.Fatalf("second reveal: %v", err)
	}
	if again.RevealedAt == nil || !again.RevealedAt.Equal(first) {
		t.Fatalf("revealed_at moved on a second reveal: %v then %v", first, again.RevealedAt)
	}

	// A step under the wrong engagement is indistinguishable from missing.
	if _, err := d.svc.RevealStep(ctx, d.actor(), d.other.ID, d.step.ID); err == nil {
		t.Fatal("cross-engagement reveal was accepted")
	} else {
		notFound(t, err)
	}

	// The revealed step is what blue's blind-filtered listing now sees.
	list, err := d.svc.ListEngagementSteps(ctx, d.eng.ID, blind.Scope{})
	if err != nil {
		t.Fatalf("list engagement steps: %v", err)
	}
	if len(list) != 1 || list[0].RevealedAt == nil {
		t.Fatalf("revealed step not visible: %+v", list)
	}
}

func TestDeleteStepCrossEngagementIsNotFound(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	err := d.svc.DeleteStep(context.Background(), d.actor(), d.other.ID, d.step.ID)
	notFound(t, err)
	if _, err := d.svc.GetStep(context.Background(), d.step.ID); err != nil {
		t.Fatalf("refused delete removed the step: %v", err)
	}
}

func TestCreateStepCrossEngagementIsNotFound(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	_, err := d.svc.CreateStep(context.Background(), d.actor(), engagement.CreateStepInput{
		EngagementID: d.other.ID,
		ScenarioID:   d.scenario.ID,
		Name:         "elsewhere",
	})
	notFound(t, err)
}
