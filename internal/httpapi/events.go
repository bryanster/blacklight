package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bryanster/blacklight/internal/authn"
	"github.com/bryanster/blacklight/internal/content"
	"github.com/bryanster/blacklight/internal/events"
	"github.com/bryanster/blacklight/internal/httpapi/apierr"
	"github.com/bryanster/blacklight/internal/httpapi/gen"
	"github.com/bryanster/blacklight/internal/store/blind"
	storecontent "github.com/bryanster/blacklight/internal/store/content"
)

// SubscribeEvents opens a text/event-stream over the shared hub (M2-004,
// extended in M4-001).
//
// The middleware gates on "authenticated session + no service token"
// (x-authz-self, security cookieSession). This handler then filters topics by
// caller permission: content.jobs* requires admin; engagement.{id} requires
// membership or admin via [handlers.topicAllowed].
//
// M4-004: Last-Event-ID triggers catch-up replay from the activity log.
// Events are replayed oldest-first then the stream transitions to live tail.
// The Allow filter drops blind-withheld events per subscriber.
func (h *handlers) SubscribeEvents(ctx context.Context, request gen.SubscribeEventsRequestObject) (gen.SubscribeEventsResponseObject, error) {
	if h.hub == nil {
		return nil, apierr.Internal(errors.New("events hub is not configured"))
	}
	cursor := ""
	if request.Params.LastEventID != nil {
		cursor = *request.Params.LastEventID
	} else if request.Params.LastEventId != nil {
		cursor = *request.Params.LastEventId
	}

	var requestedTopics []string
	if request.Params.Topics != nil {
		requestedTopics = *request.Params.Topics
	}

	// Per-topic authorization: the middleware only checks that the caller is
	// an authenticated session. This handler decides which topics they may see.
	caller, _ := authn.SubjectFrom(ctx)
	var authzTopics []string
	for _, topic := range requestedTopics {
		if h.topicAllowed(ctx, caller, topic) {
			authzTopics = append(authzTopics, topic)
		}
	}

	// Build per-topic blind scopes for engagement topics (M4-004).
	blindScopes := make(map[string]blind.Scope)
	for _, topic := range authzTopics {
		engID, ok := strings.CutPrefix(topic, events.TopicEngagementPrefix)
		if !ok {
			continue
		}
		scope, err := h.stepBlindScope(ctx, engID)
		if err != nil {
			// Engagement gone or other error — skip this topic.
			h.log.DebugContext(ctx, "events: blind scope lookup failed, skipping topic",
				slog.String("engagement_id", engID), slog.String("error", err.Error()))
			continue
		}
		blindScopes[topic] = scope
	}

	// Build the Allow filter for live events. Engagement events are
	// filtered by blind scope; content events pass through.
	allowFilter := func(ev events.Event) bool {
		// Content job events always pass.
		if ev.Topic == events.TopicContentJobs || strings.HasPrefix(ev.Topic, events.TopicContentJobs+".") {
			return true
		}
		scope, ok := blindScopes[ev.Topic]
		if !ok {
			// Unknown topic — this shouldn't happen, but don't drop.
			return true
		}
		return events.VisibleActivity(scope, ev)
	}

	// Build the Modify transform for per-subscriber event mutation.
	// Presence join/update events have focus targets stripped for blue
	// subscribers in blind engagements (M4-009).
	modifyFilter := func(ev events.Event) events.Event {
		if ev.Topic == events.TopicContentJobs || strings.HasPrefix(ev.Topic, events.TopicContentJobs+".") {
			return ev
		}
		scope, ok := blindScopes[ev.Topic]
		if !ok {
			return ev
		}
		return events.FilterPresenceEvent(ctx, scope, ev, h)
	}

	ch, unsub, err := h.hub.Subscribe(ctx, events.Subscription{
		Topics: authzTopics,
		Allow:  allowFilter,
		Modify: modifyFilter,
	})
	if err != nil {
		switch {
		case errors.Is(err, events.ErrUnknownTopic):
			return nil, apierr.Validation(apierr.Field("topics", err.Error()))
		case errors.Is(err, events.ErrNoTopics):
			return nil, apierr.Validation(apierr.Field("topics", "at least one topic is required"))
		case errors.Is(err, events.ErrTooManySubscribers):
			return nil, apierr.Internal(fmt.Errorf("event subscriber limit reached"))
		default:
			return nil, apierr.Internal(fmt.Errorf("subscribe: %w", err))
		}
	}

	pr, pw := io.Pipe()
	heartbeat := h.eventsHeartbeat
	if heartbeat <= 0 {
		heartbeat = 15 * time.Second
	}

	go h.streamWithReplay(ctx, pw, ch, unsub, heartbeat, authzTopics, blindScopes, cursor, h.log)

	cacheControl := "no-cache"
	contentType := "text/event-stream"
	return gen.SubscribeEvents200TexteventStreamResponse{
		Body: pr,
		Headers: gen.SubscribeEvents200ResponseHeaders{
			CacheControl: &cacheControl,
			ContentType:  &contentType,
		},
	}, nil
}

// topicAllowed reports whether caller may subscribe to a single topic.
// Admin sees everything. Content jobs require admin. Engagement topics
// require membership or admin.
func (h *handlers) topicAllowed(ctx context.Context, caller authn.Subject, topic string) bool {
	if string(caller.PlatformRole) == string(gen.PlatformRoleAdmin) {
		return true
	}

	if topic == events.TopicContentJobs {
		return false
	}
	if strings.HasPrefix(topic, events.TopicContentJobs+".") {
		return false
	}

	// Engagement topics: require membership.
	if rest, ok := strings.CutPrefix(topic, events.TopicEngagementPrefix); ok {
		_, isMember, err := h.ownership.Seat(ctx, rest, caller.UserID)
		if err != nil || !isMember {
			return false
		}
		return true
	}

	// Pass unknown topics through; hub's knownTopic returns 400.
	return true
}

func writeSSE(w io.Writer, ev events.Event) error {
	// id / data per the SSE spec. data is the full Event JSON (envelope).
	// No event: field — the type is in the envelope payload; all events hit
	// onmessage so the frontend parses the envelope generically (M4-003).
	var b strings.Builder
	if ev.ID != "" {
		b.WriteString("id: ")
		b.WriteString(ev.ID)
		b.WriteByte('\n')
	}
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	b.WriteString("data: ")
	b.Write(payload)
	b.WriteString("\n\n")
	_, err = io.WriteString(w, b.String())
	return err
}

// streamWithReplay replays catch-up events for each engagement topic then
// transitions to live tail. Replay is skipped when cursor is empty or
// MaxReplayEvents is 0.
func (h *handlers) streamWithReplay(
	ctx context.Context,
	w *io.PipeWriter,
	ch <-chan events.Event,
	unsub func(),
	heartbeat time.Duration,
	topics []string,
	blindScopes map[string]blind.Scope,
	cursor string,
	log *slog.Logger,
) {
	defer unsub()
	defer w.Close()

	// Leading retry hint: browsers reconnect after this many ms on drop.
	if _, err := io.WriteString(w, "retry: 3000\n\n"); err != nil {
		return
	}

	// Replay phase (M4-004): catch up for each engagement topic.
	if cursor != "" && h.eventsMaxReplay > 0 {
		for _, topic := range topics {
			engID, ok := strings.CutPrefix(topic, events.TopicEngagementPrefix)
			if !ok {
				// Content topics: no replay (Last-Event-ID ignored).
				continue
			}

			result, err := h.activity.ReplayAfter(ctx, engID, cursor, h.eventsMaxReplay)
			if err != nil {
				if log != nil {
					log.DebugContext(ctx, "events: replay failed, sending gap",
						slog.String("engagement_id", engID),
						slog.String("error", err.Error()))
				}
				// Send gap so client refetches.
				if werr := writeSSE(w, events.NewGapEvent(engID, "replay error")); werr != nil {
					return
				}
				continue
			}

			scope := blindScopes[topic]

			for _, ev := range result.Events {
				// Re-derive blind visibility at delivery time (BL-006):
				// replayed payloads carry no revealed field, so
				// VisibleActivity would fail open for step-scoped objects.
				// VisibleReplay resolves the step behind each event and
				// stamps reveal state into the events it keeps.
				kept, visible := events.VisibleReplay(ctx, scope, ev, h.resolveActivityStepID, h)
				if !visible {
					continue
				}
				if err := writeSSE(w, kept); err != nil {
					if log != nil {
						log.DebugContext(ctx, "events: replay write ended", "error", err)
					}
					return
				}
			}

			if result.Truncated {
				if err := writeSSE(w, events.NewGapEvent(engID, "replay truncated")); err != nil {
					return
				}
			}
		}
	}

	// Live tail phase.
	tick := time.NewTicker(heartbeat)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if _, err := io.WriteString(w, ": ping\n\n"); err != nil {
				return
			}
		case ev, ok := <-ch:
			if !ok {
				return
			}
			if err := writeSSE(w, ev); err != nil {
				if log != nil {
					log.DebugContext(ctx, "events: sse write ended", "error", err)
				}
				return
			}
		}
	}
}

// bridgeContentProgress maps the runner's in-process progress channel onto the
// hub. One process-lifetime goroutine; terminal statuses become
// content.job.terminal, everything else content.job.progress. Each tick is
// published on both the broadcast topic and the per-job topic.
func bridgeContentProgress(hub *events.Hub, ch <-chan content.ProgressEvent, unsub func(), log *slog.Logger) {
	defer unsub()
	for ev := range ch {
		typ := events.TypeContentJobProgress
		switch ev.Status {
		case storecontent.JobStatusSucceeded, storecontent.JobStatusFailed,
			storecontent.JobStatusCancelled, storecontent.JobStatusInterrupted:
			typ = events.TypeContentJobTerminal
		}
		data, err := json.Marshal(contentJobEventData{
			JobID:   ev.JobID,
			Phase:   ev.Phase,
			Current: ev.Current,
			Total:   ev.Total,
			Message: ev.Message,
			Status:  string(ev.Status),
		})
		if err != nil {
			if log != nil {
				log.Error("events: marshal content progress", "error", err, "job_id", ev.JobID)
			}
			continue
		}
		hub.Publish(events.TopicContentJobs, events.Event{Type: typ, Data: data})
		hub.Publish(events.TopicContentJob(ev.JobID), events.Event{Type: typ, Data: data})
	}
}

// contentJobEventData is the JSON body inside a content.job.* SSE event.
type contentJobEventData struct {
	JobID   string `json:"jobId"`
	Phase   string `json:"phase,omitempty"`
	Current int64  `json:"current"`
	Total   int64  `json:"total"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status"`
}

// eventsPath is the API path the timeout middleware must not deadline.
const eventsPath = BasePath + "/events"

// isSSEPath reports whether r is the long-lived events stream. Used by the
// timeout middleware to leave the request context unbounded.
func isSSEPath(r *http.Request) bool {
	return r.URL != nil && r.URL.Path == eventsPath
}

// IsStepRevealed implements [events.RevealLookup] for the fan-out revealed
// flag (M4-004). It returns whether the step has been revealed to the blue
// side.
func (h *handlers) IsStepRevealed(ctx context.Context, stepID string) (bool, error) {
	step, err := h.engagements.GetStep(ctx, stepID)
	if err != nil {
		return false, err
	}
	return step.RevealedAt != nil, nil
}
