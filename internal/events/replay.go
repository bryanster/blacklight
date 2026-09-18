package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bryanster/blacklight/internal/store/activity"
	"github.com/bryanster/blacklight/internal/store/blind"
)

// ReplayResult holds the events to replay before live tail begins.
type ReplayResult struct {
	// Events are the catch-up events in oldest-first order (the order they
	// happened). The caller writes them as SSE frames before the live stream.
	Events []Event

	// Truncated is true when the replay was capped (max events exceeded).
	// The caller must send a stream.gap event so the client full-refetches.
	Truncated bool
}

// ReplayAfter returns activity rows for engagementID with id strictly after
// cursor, converted to SSE Events. It returns at most maxEvents rows, oldest
// first.
//
// Cursor is an activity row id (UUIDv7). An empty cursor means "no replay" —
// returns an empty result. Activity ids are UUIDv7 and sortable by creation
// time, so `WHERE id > ? ORDER BY id ASC` gives the chronological stream
// after the cursor.
// Visibility filtering is NOT applied here — the caller re-derives blind
// visibility per event with [VisibleReplay] before sending (BL-006).
func (l *Log) ReplayAfter(ctx context.Context, engagementID, cursor string, maxEvents int) (ReplayResult, error) {
	if cursor == "" {
		return ReplayResult{}, nil
	}
	if maxEvents <= 0 {
		return ReplayResult{}, nil
	}

	// Query activity rows with id > cursor for this engagement, oldest first.
	// Limit to maxEvents+1 so we can detect truncation.
	rows, err := l.entries.ReplayAfterCursor(ctx, engagementID, cursor, maxEvents+1)
	if err != nil {
		return ReplayResult{}, fmt.Errorf("events: replay %s: %w", engagementID, err)
	}

	truncated := len(rows) > maxEvents
	if truncated {
		rows = rows[:maxEvents]
	}

	events := make([]Event, 0, len(rows))
	for _, row := range rows {
		// Build an SSE Event from the stored activity row.
		// We don't have the original Entry (with ParentIDs) for replay,
		// so we construct the data from the row fields only.
		// The parent IDs (stepId, executionId, etc.) are not available
		// in the row — they were only in the caller's Entry.
		ev := Event{
			ID:    row.ID,
			Type:  row.Verb,
			At:    row.At,
			Topic: EngagementTopic(engagementID),
			Data:  buildReplayData(row),
		}
		events = append(events, ev)
	}

	return ReplayResult{Events: events, Truncated: truncated}, nil
}

// buildReplayData constructs an SSE event payload from a stored activity row
// for replay. Unlike [buildEventData], it does not have the original Entry
// with ParentIDs — it reconstructs from what the row stores.
func buildReplayData(row activity.Row) json.RawMessage {
	m := map[string]any{
		"engagementId": row.EngagementID,
		"actorId":      row.ActorID,
		"verb":         row.Verb,
		"objectType":   row.ObjectType,
		"objectId":     row.ObjectID,
	}
	raw, err := json.Marshal(m)
	if err != nil {
		raw = []byte(`{}`)
	}
	return raw
}

// StepResolver returns the id of the step an activity object belongs to, or
// "" when the object is not step-scoped. The HTTP layer supplies the
// engagement-store-backed implementation: resolving execution, evidence and
// comment rows to their parent step needs the engagement store, which this
// package does not hold.
type StepResolver func(ctx context.Context, objectType, objectID string) (string, error)

// VisibleReplay reports whether a catch-up-replayed event is visible under
// scope, re-deriving reveal state at delivery time (BL-006).
//
// Replay rebuilds event payloads from stored activity rows, which carry no
// step linkage — buildReplayData cannot set `revealed`, so VisibleActivity
// would fail open for every step-scoped object. This helper closes that hole
// the way the activity-list filter does: resolve the step behind the event (a
// step object by its objectId, execution/evidence/comment through resolve)
// and consult lookup now, so a row for an unrevealed step is dropped and a
// step revealed after the event was written is still delivered. A kept
// step-scoped event has `revealed` stamped into its payload, so the replay
// path stops producing nil-revealed payloads and a later VisibleActivity
// pass reaches the same verdict.
//
// Failure handling follows the shared fail-open convention (see
// [Log.lookupRevealed] and the activity-list filter): an unparseable payload,
// an object the resolver cannot classify, or a resolution or lookup error
// keeps the event. A nil lookup cannot confirm anything, so step-scoped
// events are dropped — the same conservative rule as isStepVisible.
//
// A scope that withholds nothing short-circuits: events pass through
// unmodified and no lookups run.
func VisibleReplay(ctx context.Context, scope blind.Scope, ev Event, resolve StepResolver, lookup RevealLookup) (Event, bool) {
	if !scope.Withholds() {
		return ev, true
	}
	d := ParseEventData(ev.Data)
	if d == nil {
		return ev, true // unparseable — don't drop, same rule as VisibleActivity
	}
	var stepID string
	switch d.ObjectType {
	case ObjectStep:
		stepID = d.ObjectID
	case ObjectExecution, ObjectEvidence, ObjectComment:
		if resolve == nil {
			return ev, true // cannot classify without the store — fail open
		}
		resolved, err := resolve(ctx, d.ObjectType, d.ObjectID)
		if err != nil || resolved == "" {
			return ev, true // unresolvable — fail open, like the activity list
		}
		stepID = resolved
	default:
		return ev, true // engagement, member, scenario, finding, … — always visible
	}

	revealed := false
	if lookup != nil {
		ok, err := lookup.IsStepRevealed(ctx, stepID)
		if err != nil {
			return ev, true // shared convention: don't drop on lookup failure
		}
		revealed = ok
	}
	if !revealed {
		return Event{}, false
	}
	return stampRevealed(ev), true
}

// stampRevealed sets `"revealed": true` in an event payload, matching the
// shape the live fan-out produces for a revealed step-scoped object.
func stampRevealed(ev Event) Event {
	var m map[string]any
	if err := json.Unmarshal(ev.Data, &m); err != nil {
		return ev
	}
	m["revealed"] = true
	raw, err := json.Marshal(m)
	if err != nil {
		return ev
	}
	ev.Data = raw
	return ev
}

// GapEventData is the JSON payload for a stream.gap / sync.required event.
type GapEventData struct {
	EngagementID string `json:"engagementId"`
	Reason       string `json:"reason"`
}

// NewGapEvent creates a synthetic stream.gap event for an engagement.
func NewGapEvent(engagementID, reason string) Event {
	data, err := json.Marshal(GapEventData{
		EngagementID: engagementID,
		Reason:       reason,
	})
	if err != nil {
		data = []byte(`{"engagementId":"","reason":"marshal error"}`)
	}
	return Event{
		ID:    "", // synthetic — no activity row
		Type:  TypeStreamGap,
		At:    time.Now().UTC(),
		Topic: EngagementTopic(engagementID),
		Data:  data,
	}
}
