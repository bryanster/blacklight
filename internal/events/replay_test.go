package events

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/bryanster/blacklight/internal/store/activity"
	"github.com/bryanster/blacklight/internal/store/blind"
)

// The payloads here are exactly what the replay path produces: buildReplayData
// over a stored activity row. Those rows carry no step linkage, so the payload
// has no revealed field and no parent ids — which is why replay must re-derive
// reveal state with VisibleReplay instead of VisibleActivity (BL-006).

func replayEvent(row activity.Row) Event {
	return Event{
		ID:    row.ID,
		Type:  row.Verb,
		Topic: EngagementTopic(row.EngagementID),
		Data:  buildReplayData(row),
	}
}

// countingResolver answers with the configured step id (or error) and counts
// calls, so tests can assert a lookup did or did not run.
type countingResolver struct {
	stepID string
	err    error
	calls  int
}

func (s *countingResolver) resolveStep(_ context.Context, _, _ string) (string, error) {
	s.calls++
	if s.err != nil {
		return "", s.err
	}
	return s.stepID, nil
}

type countingLookup struct {
	revealed bool
	err      error
	calls    int
}

func (s *countingLookup) IsStepRevealed(_ context.Context, _ string) (bool, error) {
	s.calls++
	return s.revealed, s.err
}

func TestVisibleReplayDropsUnrevealedStepScopedPayloads(t *testing.T) {
	blue := blind.Scope{Blind: true, Seat: "blue"}

	// Every step-scoped objectType, one row each — the shapes buildReplayData
	// emits for step, execution, evidence and comment activity rows.
	rows := []activity.Row{
		{ID: "a1", EngagementID: "eng-1", ActorID: "red-1", Verb: "step.created", ObjectType: ObjectStep, ObjectID: "step-1"},
		{ID: "a2", EngagementID: "eng-1", ActorID: "red-1", Verb: "execution.red_updated", ObjectType: ObjectExecution, ObjectID: "exec-1"},
		{ID: "a3", EngagementID: "eng-1", ActorID: "red-1", Verb: "evidence.uploaded", ObjectType: ObjectEvidence, ObjectID: "ev-1"},
		{ID: "a4", EngagementID: "eng-1", ActorID: "red-1", Verb: "comment.created", ObjectType: ObjectComment, ObjectID: "c-1"},
	}

	for _, row := range rows {
		for _, tc := range []struct {
			name     string
			revealed bool
		}{
			{"unrevealed", false},
			{"revealed", true},
		} {
			t.Run(row.Verb+"/"+tc.name, func(t *testing.T) {
				lookup := &countingLookup{revealed: tc.revealed}
				// execution/evidence/comment resolve to their parent step
				// through the resolver; step rows carry the step id in
				// objectId and never reach it.
				resolver := &countingResolver{stepID: "step-1"}

				kept, visible := VisibleReplay(context.Background(), blue, replayEvent(row), resolver.resolveStep, lookup)

				if !tc.revealed {
					if visible {
						t.Fatalf("%s for an unrevealed step was replayed to blue", row.Verb)
					}
					return
				}
				if !visible {
					t.Fatalf("%s for a revealed step was dropped", row.Verb)
				}

				// The kept event is the same frame, now carrying the reveal
				// state the stored row did not have — so a later pass through
				// the shared VisibleActivity filter reaches the same verdict
				// instead of failing open on a nil-revealed payload.
				if kept.ID != row.ID || kept.Type != row.Verb || kept.Topic != EngagementTopic(row.EngagementID) {
					t.Errorf("kept event envelope changed: %+v", kept)
				}
				if !VisibleActivity(blue, kept) {
					t.Fatalf("stamped replay payload still fails VisibleActivity for blue: %s", kept.Data)
				}
				var d map[string]any
				if err := json.Unmarshal(kept.Data, &d); err != nil {
					t.Fatalf("kept payload unparseable: %v", err)
				}
				if d["revealed"] != true {
					t.Errorf("kept payload revealed = %v, want true", d["revealed"])
				}
				if d["objectId"] != row.ObjectID {
					t.Errorf("kept payload objectId = %v, want %s", d["objectId"], row.ObjectID)
				}

				// A step object carries its own id; children must have gone
				// through the resolver to find their parent step.
				wantResolverCalls := 0
				if row.ObjectType != ObjectStep {
					wantResolverCalls = 1
				}
				if resolver.calls != wantResolverCalls {
					t.Errorf("resolver calls = %d, want %d", resolver.calls, wantResolverCalls)
				}
				if lookup.calls != 1 {
					t.Errorf("lookup calls = %d, want 1", lookup.calls)
				}
			})
		}
	}
}

func TestVisibleReplayKeepsNonStepScopedEvents(t *testing.T) {
	blue := blind.Scope{Blind: true, Seat: "blue"}
	row := activity.Row{ID: "a9", EngagementID: "eng-1", ActorID: "red-1", Verb: "finding.created", ObjectType: ObjectFinding, ObjectID: "f-1"}
	resolver := &countingResolver{stepID: "should-not-be-used"}
	lookup := &countingLookup{revealed: false}

	ev := replayEvent(row)
	kept, visible := VisibleReplay(context.Background(), blue, ev, resolver.resolveStep, lookup)

	if !visible {
		t.Fatal("non-step-scoped event was dropped for blue")
	}
	if string(kept.Data) != string(ev.Data) {
		t.Errorf("payload was rewritten: %s -> %s", ev.Data, kept.Data)
	}
	if resolver.calls != 0 || lookup.calls != 0 {
		t.Errorf("lookups ran for a non-step-scoped event (resolver %d, lookup %d)", resolver.calls, lookup.calls)
	}
}

func TestVisibleReplayShortCircuitsWithoutholdingScope(t *testing.T) {
	// Red is never held to blind mode: events pass through unmodified and no
	// lookups run, whatever the payload.
	scope := blind.Scope{Blind: true, Seat: "red"}
	row := activity.Row{ID: "a1", EngagementID: "eng-1", ActorID: "red-1", Verb: "step.created", ObjectType: ObjectStep, ObjectID: "step-1"}
	resolver := &countingResolver{stepID: "step-1"}
	lookup := &countingLookup{revealed: false}

	ev := replayEvent(row)
	kept, visible := VisibleReplay(context.Background(), scope, ev, resolver.resolveStep, lookup)

	if !visible {
		t.Fatal("event was dropped for red")
	}
	if string(kept.Data) != string(ev.Data) {
		t.Errorf("payload was rewritten for red: %s -> %s", ev.Data, kept.Data)
	}
	if lookup.calls != 0 {
		t.Errorf("lookup ran for red (%d calls)", lookup.calls)
	}
}

func TestVisibleReplayFailureHandling(t *testing.T) {
	blue := blind.Scope{Blind: true, Seat: "blue"}
	stepRow := activity.Row{ID: "a1", EngagementID: "eng-1", ActorID: "red-1", Verb: "step.created", ObjectType: ObjectStep, ObjectID: "step-1"}

	t.Run("nil lookup drops step-scoped events", func(t *testing.T) {
		// A withheld scope with no way to confirm reveal state keeps nothing:
		// the same conservative rule isStepVisible applies.
		_, visible := VisibleReplay(context.Background(), blue, replayEvent(stepRow), nil, nil)
		if visible {
			t.Fatal("nil lookup delivered a step-scoped event to blue")
		}
	})

	t.Run("lookup error keeps the event", func(t *testing.T) {
		// Shared fail-open convention: a transient store error must not
		// silently swallow catch-up events (see Log.lookupRevealed).
		lookup := &countingLookup{err: errors.New("db down")}
		kept, visible := VisibleReplay(context.Background(), blue, replayEvent(stepRow), nil, lookup)
		if !visible || string(kept.Data) == "" {
			t.Fatal("lookup error dropped a replayed event")
		}
	})

	t.Run("resolver error keeps a child event", func(t *testing.T) {
		row := activity.Row{ID: "a2", EngagementID: "eng-1", ActorID: "red-1", Verb: "execution.red_updated", ObjectType: ObjectExecution, ObjectID: "exec-1"}
		resolver := &countingResolver{err: errors.New("db down")}
		_, visible := VisibleReplay(context.Background(), blue, replayEvent(row), resolver.resolveStep, nil)
		if !visible {
			t.Fatal("resolver error dropped a replayed event")
		}
	})

	t.Run("unclassifiable object keeps the event", func(t *testing.T) {
		// An execution the store no longer knows is not classifiable; the
		// activity-list filter keeps such rows, and so does replay.
		row := activity.Row{ID: "a2", EngagementID: "eng-1", ActorID: "red-1", Verb: "execution.red_updated", ObjectType: ObjectExecution, ObjectID: "exec-1"}
		resolver := &countingResolver{stepID: ""}
		_, visible := VisibleReplay(context.Background(), blue, replayEvent(row), resolver.resolveStep, nil)
		if !visible {
			t.Fatal("unclassifiable event was dropped")
		}
	})

	t.Run("unparseable payload keeps the event", func(t *testing.T) {
		ev := Event{ID: "a1", Type: "step.created", Data: json.RawMessage(`not-json`)}
		_, visible := VisibleReplay(context.Background(), blue, ev, nil, nil)
		if !visible {
			t.Fatal("unparseable event was dropped")
		}
	})
}

func TestVisibleReplayStampMatchesLiveFanOut(t *testing.T) {
	// The stamped payload must round-trip: blue reads it as a revealed
	// step-scoped event, and the fields buildReplayData produced survive.
	blue := blind.Scope{Blind: true, Seat: "blue"}
	row := activity.Row{ID: "a1", EngagementID: "eng-1", ActorID: "red-1", Verb: "step.created", ObjectType: ObjectStep, ObjectID: "step-1"}

	kept, visible := VisibleReplay(context.Background(), blue, replayEvent(row), nil, &countingLookup{revealed: true})
	if !visible {
		t.Fatal("revealed step event was dropped")
	}
	var d EventData
	if err := json.Unmarshal(kept.Data, &d); err != nil {
		t.Fatalf("stamped payload unparseable: %v", err)
	}
	if d.Revealed == nil || !*d.Revealed {
		t.Errorf("revealed = %v, want true", d.Revealed)
	}
	if d.EngagementID != "eng-1" || d.ActorID != "red-1" || d.ObjectType != ObjectStep || d.ObjectID != "step-1" || d.Verb != "step.created" {
		t.Errorf("stamped payload lost replay fields: %+v", d)
	}
	if !strings.Contains(string(kept.Data), `"step-1"`) {
		t.Errorf("objectId missing from kept payload: %s", kept.Data)
	}
}
