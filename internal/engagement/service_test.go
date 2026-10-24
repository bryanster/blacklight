package engagement_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/bryanster/blacklight/internal/authn"
	"github.com/bryanster/blacklight/internal/authz"
	"github.com/bryanster/blacklight/internal/content/attackpin"
	"github.com/bryanster/blacklight/internal/engagement"
	"github.com/bryanster/blacklight/internal/events"
	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/store/activity"
	storecontent "github.com/bryanster/blacklight/internal/store/content"
	storengagement "github.com/bryanster/blacklight/internal/store/engagement"
	"github.com/bryanster/blacklight/internal/store/identity"
	"github.com/bryanster/blacklight/internal/store/storetest"
)

// serviceFixture is the engagement service over a real database with two
// ATT&CK versions seeded ready, so the pin assertion the create and update
// paths run has something to assert against. It deliberately does not stand up
// the content runner: AssertPinned reads the source and version rows, so two
// rows are all a pin check needs.
type serviceFixture struct {
	svc         *engagement.Service
	engagements *storengagement.Engagements
	memberships *identity.Memberships
	scenarios   *storengagement.Scenarios
	steps       *storengagement.Steps
	objects     *storecontent.Objects
	entries     *activity.Entries
	actor       authn.Subject
}

func newServiceFixture(t *testing.T) *serviceFixture {
	t.Helper()
	ctx := context.Background()

	db := storetest.Migrated(t)
	engagements := storengagement.NewEngagements(db)
	memberships := identity.NewMemberships(db)
	users := identity.NewUsers(db)
	scenarios := storengagement.NewScenarios(db)
	steps := storengagement.NewSteps(db)
	entries := activity.New(db)

	sources := storecontent.NewSources(db)
	versions := storecontent.NewVersions(db, storecontent.NewPaths(t.TempDir()))
	objects := storecontent.NewObjects(db)
	if _, err := sources.SetEnabled(ctx, storecontent.SourceIDAttack, true); err != nil {
		t.Fatalf("enable the attack source: %v", err)
	}
	seedVersion := func(version string) {
		t.Helper()
		v, err := versions.Create(ctx, storecontent.NewSourceVersion{
			SourceID: storecontent.SourceIDAttack,
			Version:  version,
			Status:   storecontent.VersionStatusReady,
		})
		if err != nil {
			t.Fatalf("create version %s: %v", version, err)
		}
		if err := versions.SetState(ctx, v.ID, storecontent.VersionStatusReady, 10, "", time.Now()); err != nil {
			t.Fatalf("mark version %s ready: %v", version, err)
		}
	}
	seedVersion("15.1")
	seedVersion("15.2")

	pin, err := attackpin.New(attackpin.Deps{
		Sources:  sources,
		Versions: versions,
		Objects:  objects,
		Paths:    storecontent.NewPaths(t.TempDir()),
	})
	if err != nil {
		t.Fatalf("attackpin.New: %v", err)
	}

	svc, err := engagement.New(engagement.Deps{
		Engagements: engagements,
		AttackPin:   pin,
		Activity:    events.New(entries),
		Memberships: memberships,
		Scenarios:   scenarios,
		Steps:       steps,
		Users:       users,
	})
	if err != nil {
		t.Fatalf("engagement.New: %v", err)
	}

	// The lead the create path seats is a real user, because
	// engagement_member.added_by is a foreign key to app.user.
	lead, err := users.Create(ctx, identity.NewUser{
		Email:        "lead@example.test",
		DisplayName:  "Lead",
		PasswordHash: "argon2id$placeholder",
		PlatformRole: authz.PlatformRoleAdmin,
		Status:       identity.StatusActive,
	})
	if err != nil {
		t.Fatalf("create the actor user: %v", err)
	}

	return &serviceFixture{
		svc:         svc,
		engagements: engagements,
		memberships: memberships,
		scenarios:   scenarios,
		steps:       steps,
		objects:     objects,
		entries:     entries,
		actor:       authn.Subject{UserID: lead.ID, Email: lead.Email, DisplayName: lead.DisplayName},
	}
}

func (d *serviceFixture) create(t *testing.T, name string) storengagement.Engagement {
	t.Helper()
	eng, err := d.svc.Create(context.Background(), d.actor, engagement.CreateInput{
		Name:          name,
		AttackVersion: "15.1",
	})
	if err != nil {
		t.Fatalf("create engagement %q: %v", name, err)
	}
	return eng
}

func (d *serviceFixture) activity(t *testing.T, engagementID string, verb events.Verb) []activity.Row {
	t.Helper()
	rows, _, err := d.entries.List(context.Background(), activity.ListFilter{
		ScopeEngagement: engagementID,
		Verb:            string(verb),
	})
	if err != nil {
		t.Fatalf("list activity: %v", err)
	}
	return rows
}

func validation(t *testing.T, err error) {
	t.Helper()
	var typed *apierr.Error
	if !errors.As(err, &typed) || typed.Status() != 400 {
		t.Fatalf("err = %v, want a 400 validation error", err)
	}
}

func TestServiceNewNamesWhatIsMissing(t *testing.T) {
	t.Parallel()

	db := storetest.Migrated(t)
	engagements := storengagement.NewEngagements(db)
	memberships := identity.NewMemberships(db)

	if _, err := engagement.New(engagement.Deps{}); err == nil || !strings.Contains(err.Error(), "no engagements") {
		t.Fatalf("New(no engagements) err = %v, want a no-engagements refusal", err)
	}
	if _, err := engagement.New(engagement.Deps{Engagements: engagements}); err == nil || !strings.Contains(err.Error(), "no memberships") {
		t.Fatalf("New(no memberships) err = %v, want a no-memberships refusal", err)
	}
	// The optional dependencies stay optional: the pair above is enough.
	if _, err := engagement.New(engagement.Deps{Engagements: engagements, Memberships: memberships}); err != nil {
		t.Fatalf("New(engagements, memberships) = %v, want nil", err)
	}
}

func TestCreateEngagementValidatesItsInput(t *testing.T) {
	t.Parallel()

	d := newServiceFixture(t)
	ctx := context.Background()

	// A name is required.
	_, err := d.svc.Create(ctx, d.actor, engagement.CreateInput{AttackVersion: "15.1"})
	validation(t, err)

	// The mode has to be one the product knows.
	_, err = d.svc.Create(ctx, d.actor, engagement.CreateInput{
		Name:          "Op",
		AttackVersion: "15.1",
		Mode:          storengagement.EngagementMode("sideways"),
	})
	validation(t, err)
}

func TestCreateEngagementDefaultsModeSeatsTheLeadAndRecordsActivity(t *testing.T) {
	t.Parallel()

	d := newServiceFixture(t)
	ctx := context.Background()

	eng, err := d.svc.Create(ctx, d.actor, engagement.CreateInput{
		Name:          "Operation Nightfall",
		AttackVersion: "15.1",
	})
	if err != nil {
		t.Fatalf("create engagement: %v", err)
	}
	if eng.Mode != storengagement.EngagementModeStandard {
		t.Errorf("mode = %q, want the standard default", eng.Mode)
	}
	if eng.CreatedBy != d.actor.UserID {
		t.Errorf("created_by = %q, want the caller", eng.CreatedBy)
	}
	if eng.Status != storengagement.EngagementStatusDraft {
		t.Errorf("status = %q, want draft", eng.Status)
	}

	// The creator is seated as lead in the same transaction as the engagement.
	members, err := d.memberships.ListByEngagement(ctx, eng.ID)
	if err != nil {
		t.Fatalf("list memberships: %v", err)
	}
	if len(members) != 1 || members[0].UserID != d.actor.UserID || members[0].Role != authz.EngagementRoleLead {
		t.Fatalf("memberships = %+v, want the caller as sole lead", members)
	}

	if rows := d.activity(t, eng.ID, events.VerbEngagementCreated); len(rows) != 1 {
		t.Fatalf("engagement.created entries = %d, want 1", len(rows))
	}
}

func TestCreateEngagementRefusesAVersionThatIsNotPinned(t *testing.T) {
	t.Parallel()

	d := newServiceFixture(t)
	ctx := context.Background()

	// Not installed.
	if _, err := d.svc.Create(ctx, d.actor, engagement.CreateInput{
		Name:          "Op",
		AttackVersion: "9.9",
	}); err == nil {
		t.Fatal("an engagement was created on an uninstalled ATT&CK version")
	}
	// Empty after normalisation is its own refusal.
	if _, err := d.svc.Create(ctx, d.actor, engagement.CreateInput{Name: "Op"}); err == nil {
		t.Fatal("an engagement was created with no ATT&CK version")
	}
}

func TestUpdateEngagementPatchesEveryFieldAndRecordsTheDelta(t *testing.T) {
	t.Parallel()

	d := newServiceFixture(t)
	ctx := context.Background()
	eng := d.create(t, "Original")

	name := "Renamed"
	client := "ACME"
	description := "the new brief"
	startsOn := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endsOn := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	pin := "15.2"
	mode := storengagement.EngagementModeBlind
	autoReveal := true

	got, err := d.svc.Update(ctx, d.actor, eng.ID, engagement.UpdateInput{
		Name:              &name,
		Client:            &client,
		Description:       &description,
		StartsOn:          &startsOn,
		EndsOn:            &endsOn,
		AttackVersion:     &pin,
		Mode:              &mode,
		AutoRevealOnStart: &autoReveal,
	})
	if err != nil {
		t.Fatalf("update engagement: %v", err)
	}
	if got.Name != name || got.Client != client || got.Description != description ||
		got.AttackVersion != pin || got.Mode != mode || !got.AutoRevealOnStart {
		t.Fatalf("engagement = %+v, want every patched field", got)
	}
	if !got.StartsOn.Equal(startsOn) || !got.EndsOn.Equal(endsOn) {
		t.Fatalf("window = %v..%v, want %v..%v", got.StartsOn, got.EndsOn, startsOn, endsOn)
	}

	// The activity delta names each changed field, and is what the feed shows.
	rows := d.activity(t, eng.ID, events.VerbEngagementUpdated)
	if len(rows) != 1 {
		t.Fatalf("engagement.updated entries = %d, want 1", len(rows))
	}
	var delta map[string]any
	if err := json.Unmarshal(rows[0].Delta, &delta); err != nil {
		t.Fatalf("unmarshal delta: %v", err)
	}
	for _, key := range []string{"name", "client", "description", "starts_on", "ends_on", "attack_version", "mode", "auto_reveal_on_start"} {
		if _, ok := delta[key]; !ok {
			t.Errorf("delta is missing %q: %v", key, delta)
		}
	}
}

func TestUpdateEngagementRejectsAnUnknownMode(t *testing.T) {
	t.Parallel()

	d := newServiceFixture(t)
	eng := d.create(t, "Op")

	bad := storengagement.EngagementMode("sideways")
	_, err := d.svc.Update(context.Background(), d.actor, eng.ID, engagement.UpdateInput{Mode: &bad})
	validation(t, err)
}

func TestUpdateEngagementFreezesThePinOnceStepsExist(t *testing.T) {
	t.Parallel()

	d := newServiceFixture(t)
	ctx := context.Background()
	eng := d.create(t, "Op")

	// With no steps, moving the pin to another installed version is allowed.
	pin := "15.2"
	moved, err := d.svc.Update(ctx, d.actor, eng.ID, engagement.UpdateInput{AttackVersion: &pin})
	if err != nil {
		t.Fatalf("move the pin with no steps: %v", err)
	}
	if moved.AttackVersion != pin {
		t.Fatalf("attack_version = %q, want %q", moved.AttackVersion, pin)
	}

	// Once a step exists the pin is frozen: the step's technique ids were
	// snapshotted against the old version.
	scenario, err := d.scenarios.Create(ctx, storengagement.NewScenario{
		EngagementID: eng.ID,
		Ordinal:      1,
		Name:         "scenario",
		Source:       storengagement.ScenarioSourceManual,
	})
	if err != nil {
		t.Fatalf("create scenario: %v", err)
	}
	if _, _, err := d.steps.CreateWithExecution(ctx, storengagement.NewStep{
		ScenarioID:    scenario.ID,
		Ordinal:       1,
		Name:          "step",
		AttackVersion: pin,
	}); err != nil {
		t.Fatalf("create step: %v", err)
	}

	back := "15.1"
	_, err = d.svc.Update(ctx, d.actor, eng.ID, engagement.UpdateInput{AttackVersion: &back})
	conflict(t, err)

	// And an uninstalled pin is refused before the freeze is even consulted.
	unknown := "9.9"
	_, err = d.svc.Update(ctx, d.actor, eng.ID, engagement.UpdateInput{AttackVersion: &unknown})
	if err == nil {
		t.Fatal("the pin moved to an uninstalled version")
	}
}

func TestDeleteEngagementRemovesItAndRecordsActivity(t *testing.T) {
	t.Parallel()

	d := newServiceFixture(t)
	ctx := context.Background()
	eng := d.create(t, "Doomed")

	if err := d.svc.Delete(ctx, d.actor, eng.ID); err != nil {
		t.Fatalf("delete engagement: %v", err)
	}
	if _, err := d.engagements.ByID(ctx, eng.ID); err == nil {
		t.Fatal("engagement still present after delete")
	} else {
		notFound(t, err)
	}
	if rows := d.activity(t, eng.ID, events.VerbEngagementDeleted); len(rows) != 1 {
		t.Fatalf("engagement.deleted entries = %d, want 1", len(rows))
	}

	// Deleting it again is a not-found, not a silent success.
	if err := d.svc.Delete(ctx, d.actor, unknownID); err == nil {
		t.Fatal("delete of an unknown engagement succeeded")
	} else {
		notFound(t, err)
	}
}
