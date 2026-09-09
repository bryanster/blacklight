package events

import (
	"context"
	"encoding/json"

	"github.com/bryanster/blacklight/internal/store/blind"
)

// EventData is the parsed inner payload of an engagement-scoped event.
// It is the JSON shape produced by [buildEventData].
type EventData struct {
	EngagementID string `json:"engagementId"`
	ActorID      string `json:"actorId"`
	Verb         string `json:"verb"`
	ObjectType   string `json:"objectType"`
	ObjectID     string `json:"objectId"`

	// Parent IDs for cache invalidation.
	ExecutionID string `json:"executionId,omitempty"`
	ScenarioID  string `json:"scenarioId,omitempty"`
	StepID      string `json:"stepId,omitempty"`

	// Revealed is whether the step this event relates to has been revealed
	// to the blue side. It is only meaningful for step-scoped objects
	// (step, execution, evidence, comment). nil means this event is not
	// step-scoped and is always visible.
	Revealed *bool `json:"revealed,omitempty"`
}

// ParseEventData unmarshals an event's Data payload into an [EventData].
// Returns nil when unmarshalling fails — the caller should treat an
// unparseable event as visible (don't drop events the server doesn't
// understand).
func ParseEventData(raw json.RawMessage) *EventData {
	var d EventData
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil
	}
	return &d
}

// VisibleActivity reports whether an engagement-scoped event is visible to
// a subscriber under blind.Scope. It is the one shared helper used by:
//
//  1. [Subscription.Allow] on live engagement streams,
//  2. catch-up replay — through [VisibleReplay], which re-derives the
//     reveal state stored activity rows do not carry (BL-006),
//  3. the activity list API (converted to SSE), and
//  4. presence focus stripping (M4-006).
//
// Synthetic events (stream.gap / sync.required) have no objectType and
// are always visible. Events whose Data cannot be parsed are visible
// (don't drop what we can't classify).
func VisibleActivity(scope blind.Scope, ev Event) bool {
	d := ParseEventData(ev.Data)
	if d == nil {
		return true // unparseable — don't drop
	}
	return visibleEventData(scope, d)
}

// visibleEventData is the parsed-fast-path half of VisibleActivity.
func visibleEventData(scope blind.Scope, d *EventData) bool {
	if !scope.Withholds() {
		return true
	}

	// Non-step-scoped objects (engagement, member, scenario, finding) are
	// always visible: blind mode withholds steps and their descendants.
	switch d.ObjectType {
	case ObjectStep, ObjectExecution, ObjectEvidence, ObjectComment:
		// These require the step to be revealed.
	default:
		return true
	}

	// Revealed is nil when the event pre-dates the M4-004 fan-out change
	// or the field was omitted. Treat nil conservatively: reveal the event
	// rather than risk dropping legitimate events. A nil revealed field
	// after this ticket's deploy is a bug; before it, it's older events.
	if d.Revealed == nil {
		return true
	}
	return *d.Revealed
}

// RevealLookup resolves whether a step or step-scoped object is revealed.
// Engagements and the HTTP handler supply an implementation backed by the
// engagement store. Nil means no lookup — the fan-out omits the revealed
// field from event data (conservative: everything visible).
type RevealLookup interface {
	// IsStepRevealed returns whether the step with this id has been revealed.
	IsStepRevealed(ctx context.Context, stepID string) (bool, error)
}

// StepIDForEvent returns the step id that an event is about, if any.
// For step objects, ObjectID is the step id. For execution/evidence/comment,
// the StepID parent field carries it.
func StepIDForEvent(d *EventData) string {
	if d == nil {
		return ""
	}
	switch d.ObjectType {
	case ObjectStep:
		return d.ObjectID
	case ObjectExecution, ObjectEvidence, ObjectComment:
		return d.StepID
	default:
		return ""
	}
}

// presenceData is the JSON shape carried by presence.join and
// presence.update events (see httpapi.presenceEventJSON).
type presenceData struct {
	EngagementID string `json:"engagementId"`
	UserID       string `json:"userId"`
	DisplayName  string `json:"displayName"`
	StepID       string `json:"stepId"`
	ExecutionID  string `json:"executionId"`
}

// FilterPresenceEvent strips unrevealed focus targets from a presence
// join or update event for the given blind scope. Leave events pass
// through unchanged. This is the per-subscriber transform wired via
// [Subscription.Modify] in the SSE path (M4-009).
//
// lookup may be nil — unrevealed focus is then dropped conservatively
// (the step id is stripped). A non-nil lookup confirms reveal status from
// the store.
func FilterPresenceEvent(ctx context.Context, scope blind.Scope, ev Event, lookup RevealLookup) Event {
	if !scope.Withholds() {
		return ev
	}
	if ev.Type != TypePresenceJoin && ev.Type != TypePresenceUpdate {
		return ev
	}

	var pd presenceData
	if err := json.Unmarshal(ev.Data, &pd); err != nil {
		return ev // unparseable — don't mutate
	}

	// If no focus is set, nothing to strip.
	if pd.StepID == "" && pd.ExecutionID == "" {
		return ev
	}

	var stripped bool
	if pd.StepID != "" && !isStepVisible(ctx, pd.StepID, lookup) {
		pd.StepID = ""
		stripped = true
	}
	if pd.ExecutionID != "" && !isStepVisible(ctx, pd.ExecutionID, lookup) {
		pd.ExecutionID = ""
		stripped = true
	}

	if !stripped {
		return ev
	}

	// Re-serialize with stripped focus.
	b, err := json.Marshal(pd)
	if err != nil {
		return ev
	}
	ev.Data = json.RawMessage(b)
	return ev
}

// isStepVisible reports whether a step id should be visible.
// When lookup is nil we conservatively strip (return false).
func isStepVisible(ctx context.Context, stepID string, lookup RevealLookup) bool {
	if lookup == nil {
		return false
	}
	revealed, err := lookup.IsStepRevealed(ctx, stepID)
	if err != nil || !revealed {
		return false
	}
	return true
}
