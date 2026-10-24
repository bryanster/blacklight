package engagement_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/bryanster/blacklight/internal/authn"
	"github.com/bryanster/blacklight/internal/engagement"
	"github.com/bryanster/blacklight/internal/events"
	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/store/activity"
	"github.com/bryanster/blacklight/internal/store/blind"
	storengagement "github.com/bryanster/blacklight/internal/store/engagement"
	"github.com/bryanster/blacklight/internal/store/identity"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

// workbook is the domain service over a real database with one seeded
// engagement → scenario → step → execution → comment chain, plus a second
// engagement to drive the cross-engagement bindings. It exists so the workbook
// CRUD paths (scenarios, steps, findings, comments, imports) can be exercised
// through the service rather than the store — the store's own behaviour has its
// own tests under internal/store/engagement.
//
// Everything is built through the repositories rather than the service so a
// test starts from a known ordinal instead of whatever the service happened to
// assign.
type workbook struct {
	svc         *engagement.Service
	engagements *storengagement.Engagements
	scenarios   *storengagement.Scenarios
	steps       *storengagement.Steps
	executions  *storengagement.Executions
	findings    *storengagement.Findings
	comments    *storengagement.Comments
	entries     *activity.Entries

	eng      storengagement.Engagement
	other    storengagement.Engagement
	scenario storengagement.Scenario
	step     storengagement.Step
	exec     storengagement.Execution
	comment  storengagement.Comment
}

const workbookActorID = "0192f1a0-0000-7000-8000-0000000000ff"

// unknownID is a well-formed id that resolves to nothing, for the "missing
// resource" half of every binding.
const unknownID = "0192f1a0-0000-7000-8000-0000000000aa"

func newWorkbook(t *testing.T) *workbook {
	t.Helper()
	ctx := context.Background()

	db := storetest.Migrated(t)
	engagements := storengagement.NewEngagements(db)
	scenarios := storengagement.NewScenarios(db)
	steps := storengagement.NewSteps(db)
	executions := storengagement.NewExecutions(db)
	findings := storengagement.NewFindings(db)
	comments := storengagement.NewComments(db)
	entries := activity.New(db)

	// Wired the way httpapi.NewServer wires it: an activity log attached, so the
	// record*Activity paths run here as well rather than only in the API tests.
	svc, err := engagement.New(engagement.Deps{
		Engagements: engagements,
		Activity:    events.New(entries),
		Memberships: identity.NewMemberships(db),
		Scenarios:   scenarios,
		Steps:       steps,
		Executions:  executions,
		Comments:    comments,
		Findings:    findings,
	})
	if err != nil {
		t.Fatalf("engagement.New: %v", err)
	}

	newEngagement := func(name string) storengagement.Engagement {
		eng, err := engagements.Create(ctx, storengagement.NewEngagement{
			Name:          name,
			AttackVersion: "15.1",
			Mode:          storengagement.EngagementModeStandard,
			CreatedBy:     workbookActorID,
		})
		if err != nil {
			t.Fatalf("create engagement %s: %v", name, err)
		}
		return eng
	}

	eng := newEngagement("Workbook")
	other := newEngagement("Other")

	scenario, err := scenarios.Create(ctx, storengagement.NewScenario{
		EngagementID: eng.ID,
		Ordinal:      1,
		Name:         "scenario-1",
		Source:       storengagement.ScenarioSourceManual,
	})
	if err != nil {
		t.Fatalf("create scenario: %v", err)
	}

	step, exec, err := steps.CreateWithExecution(ctx, storengagement.NewStep{
		ScenarioID:    scenario.ID,
		Ordinal:       1,
		Name:          "step-1",
		TechniqueID:   "T1059",
		AttackVersion: "15.1",
	})
	if err != nil {
		t.Fatalf("create step: %v", err)
	}

	comment, err := comments.Create(ctx, storengagement.NewComment{
		ExecutionID: exec.ID,
		AuthorID:    workbookActorID,
		Body:        "first",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	return &workbook{
		svc:         svc,
		engagements: engagements,
		scenarios:   scenarios,
		steps:       steps,
		executions:  executions,
		findings:    findings,
		comments:    comments,
		entries:     entries,
		eng:         eng,
		other:       other,
		scenario:    scenario,
		step:        step,
		exec:        exec,
		comment:     comment,
	}
}

func (d *workbook) actor() authn.Subject {
	return authn.Subject{UserID: workbookActorID, Email: "actor@example.test", DisplayName: "Actor"}
}

// closeEngagement moves the engagement to closed, the state every write path
// except a finding status change refuses.
func (d *workbook) closeEngagement(t *testing.T, id string) {
	t.Helper()
	if _, err := d.svc.SetStatus(context.Background(), d.actor(), id, storengagement.EngagementStatusClosed); err != nil {
		t.Fatalf("close engagement: %v", err)
	}
}

func (d *workbook) addScenario(t *testing.T, engID, name string, ordinal int) storengagement.Scenario {
	t.Helper()
	sc, err := d.scenarios.Create(context.Background(), storengagement.NewScenario{
		EngagementID: engID,
		Ordinal:      ordinal,
		Name:         name,
		Source:       storengagement.ScenarioSourceManual,
	})
	if err != nil {
		t.Fatalf("seed scenario %s: %v", name, err)
	}
	return sc
}

func (d *workbook) addStep(t *testing.T, scenarioID, name string, ordinal int) storengagement.Step {
	t.Helper()
	step, _, err := d.steps.CreateWithExecution(context.Background(), storengagement.NewStep{
		ScenarioID:    scenarioID,
		Ordinal:       ordinal,
		Name:          name,
		AttackVersion: "15.1",
	})
	if err != nil {
		t.Fatalf("seed step %s: %v", name, err)
	}
	return step
}

func (d *workbook) createFinding(t *testing.T, title string) storengagement.Finding {
	t.Helper()
	f, err := d.svc.CreateFinding(context.Background(), d.actor(), engagement.CreateFindingInput{
		EngagementID: d.eng.ID,
		Title:        title,
	})
	if err != nil {
		t.Fatalf("create finding %s: %v", title, err)
	}
	return f
}

func conflict(t *testing.T, err error) {
	t.Helper()
	var typed *apierr.Error
	if !errors.As(err, &typed) || typed.Status() != 409 {
		t.Fatalf("err = %v, want a 409 conflict", err)
	}
}

// notFound accepts either shape a missing row arrives in: the repositories that
// map sql.ErrNoRows to apierr.NotFound (scenarios) and the ones that let it
// through for httpapi's ownership layer to normalise (steps, findings,
// comments). See internal/httpapi/ownership.go, which treats the two the same.
func notFound(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, apierr.ErrNotFound) && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want a not-found", err)
	}
}

// --- Scenarios -----------------------------------------------------------------

func TestCreateScenarioDefaultsSourceAndAppendsOrdinal(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	// An empty source is not a choice the caller has to make: a scenario typed
	// by hand is "manual".
	second, err := d.svc.CreateScenario(ctx, d.actor(), engagement.CreateScenarioInput{
		EngagementID: d.eng.ID,
		Name:         "second",
	})
	if err != nil {
		t.Fatalf("create scenario: %v", err)
	}
	if second.Source != storengagement.ScenarioSourceManual {
		t.Errorf("source = %q, want %q", second.Source, storengagement.ScenarioSourceManual)
	}
	if second.Ordinal != 2 {
		t.Errorf("ordinal = %d, want 2 (the next dense position)", second.Ordinal)
	}

	third, err := d.svc.CreateScenario(ctx, d.actor(), engagement.CreateScenarioInput{
		EngagementID: d.eng.ID,
		Name:         "third",
	})
	if err != nil {
		t.Fatalf("create third scenario: %v", err)
	}
	if third.Ordinal != 3 {
		t.Errorf("ordinal = %d, want 3", third.Ordinal)
	}

	// The creation is an activity entry, which is what the feed reads.
	rows, _, err := d.entries.List(ctx, activity.ListFilter{
		ScopeEngagement: d.eng.ID,
		Verb:            string(events.VerbScenarioCreated),
	})
	if err != nil {
		t.Fatalf("list activity: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("scenario.created entries = %d, want 2", len(rows))
	}
}

func TestCreateScenarioRefusesClosedEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	d.closeEngagement(t, d.eng.ID)

	_, err := d.svc.CreateScenario(context.Background(), d.actor(), engagement.CreateScenarioInput{
		EngagementID: d.eng.ID,
		Name:         "too late",
	})
	conflict(t, err)
}

func TestGetScenarioRoundTripsAndMissingIsNotFound(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	got, err := d.svc.GetScenario(ctx, d.scenario.ID)
	if err != nil {
		t.Fatalf("get scenario: %v", err)
	}
	if got.ID != d.scenario.ID || got.Name != "scenario-1" || got.EngagementID != d.eng.ID {
		t.Fatalf("scenario = %+v, want the seeded one", got)
	}

	if _, err := d.svc.GetScenario(ctx, unknownID); !errors.Is(err, apierr.ErrNotFound) {
		t.Fatalf("missing scenario: err = %v, want not found", err)
	}
}

func TestReorderScenariosRejectsAnIncompleteOrForeignSet(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	d.addScenario(t, d.eng.ID, "a", 2)
	foreign := d.addScenario(t, d.other.ID, "foreign", 1)

	// One scenario short of the set: the request would leave an ordinal
	// undefined.
	_, err := d.svc.ReorderScenarios(ctx, d.actor(), d.eng.ID, []string{d.scenario.ID})
	conflict(t, err)

	// Right count, wrong engagement: a foreign id is not silently ignored.
	_, err = d.svc.ReorderScenarios(ctx, d.actor(), d.eng.ID, []string{d.scenario.ID, foreign.ID})
	conflict(t, err)

	// A refused reorder changes nothing.
	list, err := d.svc.ListScenarios(ctx, d.eng.ID)
	if err != nil {
		t.Fatalf("list scenarios: %v", err)
	}
	if len(list) != 2 || list[0].ID != d.scenario.ID || list[0].Ordinal != 1 {
		t.Fatalf("refused reorder mutated the ordinals: %+v", list)
	}
}

func TestDeleteScenarioRemovesItsChildGraph(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()

	if err := d.svc.DeleteScenario(ctx, d.actor(), d.eng.ID, d.scenario.ID); err != nil {
		t.Fatalf("delete scenario: %v", err)
	}

	if _, err := d.scenarios.ByID(ctx, d.scenario.ID); !errors.Is(err, apierr.ErrNotFound) {
		t.Errorf("scenario still present after delete: %v", err)
	}
	// The step (and its execution and comment) go with it.
	if _, err := d.steps.ByID(ctx, d.step.ID); err == nil {
		t.Errorf("step still present after deleting its scenario: %v", err)
	} else {
		notFound(t, err)
	}
	if _, err := d.comments.ByID(ctx, d.comment.ID); err == nil {
		t.Errorf("comment still present after deleting its scenario: %v", err)
	} else {
		notFound(t, err)
	}

	list, err := d.svc.ListScenarios(ctx, d.eng.ID)
	if err != nil {
		t.Fatalf("list scenarios: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("scenarios after delete = %d, want 0", len(list))
	}
}

func TestDeleteScenarioRefusesClosedEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	d.closeEngagement(t, d.eng.ID)

	conflict(t, d.svc.DeleteScenario(context.Background(), d.actor(), d.eng.ID, d.scenario.ID))
}

func TestDeleteScenarioCrossEngagementIsNotFound(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	err := d.svc.DeleteScenario(context.Background(), d.actor(), d.other.ID, d.scenario.ID)
	if !errors.Is(err, apierr.ErrNotFound) {
		t.Fatalf("cross-engagement delete: err = %v, want not found", err)
	}
	// And the scenario it named is untouched.
	if _, err := d.scenarios.ByID(context.Background(), d.scenario.ID); err != nil {
		t.Fatalf("refused delete removed the scenario: %v", err)
	}
}

// --- Steps ---------------------------------------------------------------------

func TestCreateStepAppendsOrdinalAndDefaultsJSONColumns(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	step, err := d.svc.CreateStep(context.Background(), d.actor(), engagement.CreateStepInput{
		EngagementID: d.eng.ID,
		ScenarioID:   d.scenario.ID,
		Name:         "second",
	})
	if err != nil {
		t.Fatalf("create step: %v", err)
	}
	if step.Ordinal != 2 {
		t.Errorf("ordinal = %d, want 2", step.Ordinal)
	}
	// A step created without procedure/tools/controls gets empty JSON
	// containers, not NULLs a renderer would have to guard against.
	if string(step.Procedure) != "{}" {
		t.Errorf("procedure = %s, want {}", step.Procedure)
	}
	if string(step.Tools) != "[]" {
		t.Errorf("tools = %s, want []", step.Tools)
	}
	if string(step.ControlsInScope) != "[]" {
		t.Errorf("controls_in_scope = %s, want []", step.ControlsInScope)
	}
	if step.AttackVersion != "15.1" {
		t.Errorf("attack_version = %q, want the engagement's pin", step.AttackVersion)
	}
}

func TestCreateStepRefusesClosedEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	d.closeEngagement(t, d.eng.ID)

	_, err := d.svc.CreateStep(context.Background(), d.actor(), engagement.CreateStepInput{
		EngagementID: d.eng.ID,
		ScenarioID:   d.scenario.ID,
		Name:         "too late",
	})
	conflict(t, err)
}

func TestListEngagementStepsSpansEveryScenario(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	sc2 := d.addScenario(t, d.eng.ID, "scenario-2", 2)
	step2 := d.addStep(t, sc2.ID, "step-2", 1)

	got, err := d.svc.ListEngagementSteps(ctx, d.eng.ID, blind.Scope{})
	if err != nil {
		t.Fatalf("list engagement steps: %v", err)
	}
	seen := map[string]bool{}
	for _, s := range got {
		seen[s.ID] = true
	}
	if len(got) != 2 || !seen[d.step.ID] || !seen[step2.ID] {
		t.Fatalf("steps = %+v, want both scenarios' steps", got)
	}
}

func TestReorderStepsRejectsAnIncompleteOrForeignSet(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	ctx := context.Background()
	d.addStep(t, d.scenario.ID, "step-2", 2)

	// One step short of the set.
	_, err := d.svc.ReorderSteps(ctx, d.actor(), d.eng.ID, d.scenario.ID, []string{d.step.ID})
	conflict(t, err)

	// A step from another scenario in the same engagement is still foreign to
	// this one.
	sc2 := d.addScenario(t, d.eng.ID, "scenario-2", 2)
	foreign := d.addStep(t, sc2.ID, "elsewhere", 1)
	_, err = d.svc.ReorderSteps(ctx, d.actor(), d.eng.ID, d.scenario.ID, []string{d.step.ID, foreign.ID})
	conflict(t, err)

	// The refused reorder left the scenario's ordinals alone.
	list, err := d.svc.ListSteps(ctx, d.eng.ID, d.scenario.ID, blind.Scope{})
	if err != nil {
		t.Fatalf("list steps: %v", err)
	}
	if len(list) != 2 || list[0].ID != d.step.ID || list[1].Ordinal != 2 {
		t.Fatalf("refused reorder mutated the ordinals: %+v", list)
	}
}

func TestDeleteStepRefusesClosedEngagement(t *testing.T) {
	t.Parallel()

	d := newWorkbook(t)
	d.closeEngagement(t, d.eng.ID)

	conflict(t, d.svc.DeleteStep(context.Background(), d.actor(), d.eng.ID, d.step.ID))
}
