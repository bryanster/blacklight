package engagement_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bryanster/blacklight/internal/engagement"
	storecontent "github.com/bryanster/blacklight/internal/store/content"
	storengagement "github.com/bryanster/blacklight/internal/store/engagement"
)

func TestImportPlanCreatesScenarioAndPendingSteps(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	res, err := d.svc.ImportPlan(ctx, d.actor(), engagement.ImportPlanInput{
		EngagementID:   d.eng.ID,
		PlanName:       "CTID emulation plan",
		PlanNarrative:  "the adversary's story",
		AdversaryName:  "APT999",
		PlanExternalID: "plan-42",
		PlanID:         "catalog-42",
		Steps: []engagement.ImportPlanStepInput{
			{
				Name:        "one",
				Description: "first step",
				Procedure:   json.RawMessage(`{"executor":"sh","command":"whoami"}`),
			},
			{Name: "two"},
		},
	})
	if err != nil {
		t.Fatalf("import plan: %v", err)
	}

	// The scenario carries the plan's provenance.
	if res.Scenario.Name != "CTID emulation plan" ||
		res.Scenario.Source != storengagement.ScenarioSourceCTID ||
		res.Scenario.SourceRef != "plan-42" ||
		res.Scenario.ThreatActor != "APT999" ||
		res.Scenario.EngagementID != d.eng.ID {
		t.Fatalf("scenario = %+v, want the plan's provenance", res.Scenario)
	}

	if len(res.Steps) != 2 {
		t.Fatalf("imported %d steps, want 2", len(res.Steps))
	}
	if res.Steps[0].Ordinal != 1 || res.Steps[1].Ordinal != 2 {
		t.Fatalf("ordinals = %d, %d, want 1, 2", res.Steps[0].Ordinal, res.Steps[1].Ordinal)
	}

	// Each step arrives with a pending execution, which is what makes it
	// runnable in the workbook.
	for _, step := range res.Steps {
		exec, err := d.executions.ByStepID(ctx, step.ID)
		if err != nil {
			t.Fatalf("execution for step %s: %v", step.ID, err)
		}
		if exec.Status != storengagement.ExecutionStatusPending {
			t.Errorf("execution status = %q, want pending", exec.Status)
		}
	}

	// The description is folded into the procedure JSON so a renderer has it.
	var proc map[string]any
	if err := json.Unmarshal(res.Steps[0].Procedure, &proc); err != nil {
		t.Fatalf("unmarshal procedure: %v", err)
	}
	if proc["description"] != "first step" || proc["name"] != "one" || proc["executor"] != "sh" {
		t.Fatalf("procedure = %s, want the description merged into the plan's payload", res.Steps[0].Procedure)
	}

	// With no attackpin service wired, technique resolution is skipped rather
	// than reported as a warning.
	if len(res.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want none", res.Warnings)
	}
}

func TestImportPlanHonoursNameOverrideAndStartingOrdinal(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	res, err := d.svc.ImportPlan(context.Background(), d.actor(), engagement.ImportPlanInput{
		EngagementID:    d.eng.ID,
		PlanName:        "original name",
		NameOverride:    "overridden name",
		PlanExternalID:  "plan-7",
		StartingOrdinal: 5,
		Steps: []engagement.ImportPlanStepInput{
			{Name: "one"},
			{Name: "two"},
		},
	})
	if err != nil {
		t.Fatalf("import plan: %v", err)
	}
	if res.Scenario.Name != "overridden name" {
		t.Errorf("scenario name = %q, want the override", res.Scenario.Name)
	}
	if res.Steps[0].Ordinal != 5 || res.Steps[1].Ordinal != 6 {
		t.Errorf("ordinals = %d, %d, want 5, 6", res.Steps[0].Ordinal, res.Steps[1].Ordinal)
	}
}

func TestImportPlanRefusesClosedEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	d.closeEngagement(t, d.eng.ID)

	_, err := d.svc.ImportPlan(context.Background(), d.actor(), engagement.ImportPlanInput{
		EngagementID:   d.eng.ID,
		PlanName:       "too late",
		PlanExternalID: "plan-1",
		Steps:          []engagement.ImportPlanStepInput{{Name: "one"}},
	})
	conflict(t, err)
}

// With a pinned ATT&CK version wired — the shape the server has and the
// attackpin-less fixture above does not — a step's technique is resolved
// against the pin, and a technique that is not in it becomes a warning rather
// than a failed import.
func TestImportPlanResolvesTechniquesAndWarnsOnTheRest(t *testing.T) {
	t.Parallel()

	d := newServiceFixture(t)
	ctx := context.Background()
	eng := d.create(t, "Import target")

	if _, err := d.objects.CreateTechnique(ctx, storecontent.Technique{
		SourceID:    storecontent.SourceIDAttack,
		Version:     "15.1",
		ExternalID:  "T1059",
		Name:        "Command and Scripting Interpreter",
		Description: "execution via a shell",
	}); err != nil {
		t.Fatalf("seed technique: %v", err)
	}

	res, err := d.svc.ImportPlan(ctx, d.actor, engagement.ImportPlanInput{
		EngagementID:   eng.ID,
		PlanName:       "Plan with techniques",
		PlanExternalID: "plan-tech",
		Steps: []engagement.ImportPlanStepInput{
			{Name: "resolved", TechniqueExternalID: "T1059"},
			{Name: "missing", TechniqueExternalID: "T9999"},
		},
	})
	if err != nil {
		t.Fatalf("import plan: %v", err)
	}

	if got := res.Steps[0].TechniqueID; got != "T1059" {
		t.Errorf("resolved step technique = %q, want T1059", got)
	}
	if got := res.Steps[1].TechniqueID; got != "" {
		t.Errorf("unresolved step technique = %q, want empty", got)
	}
	if len(res.Warnings) != 1 {
		t.Fatalf("warnings = %+v, want one for the unresolved technique", res.Warnings)
	}
	warn := res.Warnings[0]
	if warn.StepOrdinal != 2 || warn.TechniqueExternalID != "T9999" || !strings.Contains(warn.Message, "not found") {
		t.Fatalf("warning = %+v, want the second step's unresolved T9999", warn)
	}
}
