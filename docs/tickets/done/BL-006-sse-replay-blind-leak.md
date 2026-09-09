# BL-006 — SSE catch-up replay delivers unrevealed-step events to blue

**Severity:** Medium · **Status:** closed (2026-09-09) · **Proven:** 2026-09-09

## Why

In a blind engagement the blue seat must not learn that an unrevealed step exists
(`docs/authz.md` guards; M4-004, M4-009). Three delivery paths honour that:

1. **Live SSE** — fan-out stamps `revealed` into every step-scoped event payload
   (`internal/events/activity.go` `buildEventData`, via `RevealLookup`) and the subscriber's
   `Allow` filter drops `revealed: false` events for the blue seat.
2. **Activity list** — `filterBlindActivity` (`internal/httpapi/activityhandlers.go`) resolves the
   step behind each row and drops unrevealed ones.
3. **Archive export** — scope-filtered step, execution, comment and finding-link data.

The **catch-up replay path does not.** On reconnect with `Last-Event-ID`
(`GET /api/v1/events?topics=engagement.{id}&lastEventId={cursor}`), `streamWithReplay` fetches raw
activity rows and rebuilds event payloads with `buildReplayData`, which **cannot set `revealed`**
— the stored rows carry no step linkage. `visibleEventData` then treats a nil `revealed` as
**visible** ("Treat nil conservatively: reveal the event rather than risk dropping legitimate
events"). Every replayed step-scoped event therefore fails open.

A blue member who reconnects with any stale cursor receives `step.created` / `step.updated` /
`execution.created` / `comment.created` / `evidence.uploaded` events for **unrevealed steps** —
objectId, actorId and timestamp included — while the direct read of the same step is (correctly) a
404 and the live stream would have dropped the identical event. The leak is exactly what blind
mode exists to withhold: that a hidden step exists, that red is operating against it, how fast,
and who.

## Proof (reproduced against the real chain, 2026-09-09)

Integration repro on the standard `httpapi` harness (`newAuthServer` → real chi chain → real
migrated DuckDB → real SSE hub → `streamWithReplay`):

1. Seed a blind engagement with red and blue members. Red creates a step via
   `POST /engagements/{id}/scenarios/{sid}/steps` and patches it — activity rows are recorded
   (`step.created`, `step.updated`), step unrevealed.
2. Blue `GET` on the unrevealed step → **404** (the middleware blind guard holds).
3. Blue subscribes with a stale cursor:
   `GET /api/v1/events?topics=engagement.{engId}&lastEventId=00000000-0000-7000-8000-000000000000`
   (a zero UUIDv7 sorts before all real activity ids; the query twin of `Last-Event-ID`).
4. Replay delivers both hidden events:

```
REPRO CONFIRMED: blue replay received 2 events about unrevealed step
01a087fd-bdd7-7694-a8ab-97cdc1781890:
  step.created objectId=01a087fd-bdd7-7694-a8ab-97cdc1781890
  step.updated objectId=01a087fd-bdd7-7694-a8ab-97cdc1781890
```

Note: `testConfig` (`internal/httpapi/httpapi_test.go`) leaves `Events.MaxReplayEvents` at 0, and
replay is skipped when it is 0 — production defaults it to 500 (`BLACKLIGHT_EVENTS_MAX_REPLAY`,
`internal/config/config.go`). The harness therefore never exercised replay, which is why no
existing test caught this.

## Sink

| Where | What |
|---|---|
| `internal/events/replay.go` (`buildReplayData`) | rebuilds payloads from stored rows only; never sets `revealed` (nor `stepId`/`executionId`) |
| `internal/events/blind.go` (`visibleEventData`) | nil `revealed` on a step-scoped object ⇒ **visible** (fail-open, by comment) |
| `internal/httpapi/events.go` (`streamWithReplay`) | replays rows and filters with `VisibleActivity` over those nil-`revealed` payloads |
| Contrast, correct: `internal/events/activity.go` | live fan-out stamps `revealed` from `RevealLookup` |
| Contrast, correct: `internal/httpapi/activityhandlers.go` | activity list resolves the step behind each row and drops unrevealed |

## Scope

**In**

- Replay path re-derives reveal state before delivery. Preferred shape mirrors
  `filterBlindActivity`: resolve the step behind each replayed row (step → `object_id`;
  execution/evidence/comment → parent step, like `resolveActivityStepID`) and consult
  `RevealLookup` at delivery time, so a step revealed after the event was written is not
  over-withheld and a row for an unrevealed step is dropped. Either filter inside
  `ReplayAfter` (passing a lookup in) or in `streamWithReplay` before `writeSSE`.
- Regression tests: one at the `httpapi` layer reproducing the proof above (blue, stale
  `lastEventId`, assert no frame names the hidden step id; assert revealed-step and
  non-step-scoped events still replay), one unit test in `internal/events` covering
  `buildReplayData`-shaped payloads under `VisibleActivity` with a blue blind scope.

**Out**

- The live-path filter, presence focus stripping, the activity list, the archive export — all
  already correct.
- The nil-⇒-visible default in `visibleEventData` for *legacy* fan-out payloads (pre-M4-004
  rows); keep tolerance there, but the replay path must stop *producing* nil-`revealed`
  payloads for step-scoped objects.
- M5-002 (query-layer blind fence for step reads) — separate, already on the backlog.
- BL-007 (findings-list leak) — separate ticket, same family.

## Files

- `internal/events/replay.go` — lookup-aware replay (new `ReplayAfter` parameter or wrapper).
- `internal/httpapi/events.go` — pass `h` (the `RevealLookup`) into the replay filter.
- `internal/events/replay_test.go` (new) and a regression in `internal/httpapi` next to
  `blind_integration_test.go`.

## Acceptance criteria

- [x] Blue in a blind engagement reconnecting with `Last-Event-ID` receives **no** event naming an
      unrevealed step — directly (`step.*`) or via parent linkage (`execution.*`, `comment.*`,
      `evidence.*`).
- [x] Blue still receives replayed events for revealed steps, non-step-scoped events
      (engagement/member/scenario/finding), and synthetic gap events.
- [x] A step revealed *between* event write and replay is delivered (live-status lookup, not
      write-time snapshot).
- [x] `MaxReplayEvents = 0` still disables replay; truncation still emits `stream.gap`.
- [x] Red/lead/admin replay behaviour unchanged.

## Tests

The proof above becomes the regression: same harness, same requests, assert absence of the hidden
step id in every replayed frame instead of its presence. `testConfig` must set
`Events.MaxReplayEvents` (e.g. 500) or the test exercises nothing — say so in a comment so the
next reader does not re-derive it.

## Implementation notes

Landed 2026-09-09. The shape follows the "filter at delivery" option, but as a standalone
helper rather than a `ReplayAfter` parameter:

- `internal/events/replay.go` gains `StepResolver` (func type: objectType+objectID → parent
  step id) and `VisibleReplay(ctx, scope, ev, resolve, lookup) (Event, bool)`. It mirrors
  `filterBlindActivity`'s semantics — resolve the step behind each row (step → objectId;
  execution/evidence/comment → resolver), consult `RevealLookup` live — and *stamps*
  `revealed: true` into the payloads it keeps, so the replay path stops producing
  nil-`revealed` payloads and a later `VisibleActivity` pass agrees with the replay verdict.
  `ReplayAfter` itself is unchanged (still unfiltered); filtering happens in
  `streamWithReplay` next to the `writeSSE` loop, where `scope`, `ctx` and the handler live.
- `streamWithReplay` calls `events.VisibleReplay(ctx, scope, ev, h.resolveActivityStepID, h)`
  per event; the resolver is the same helper the activity list uses, so both paths resolve
  parent linkage identically.
- Failure semantics (documented on `VisibleReplay`): unparseable payload, unclassifiable
  object (resolver error or ""), or lookup error ⇒ keep the event — the shared fail-open
  convention (`Log.lookupRevealed`, activity list). A withheld scope with a nil lookup drops
  step-scoped events — the same conservative rule as `isStepVisible`.
- `testConfig` (`internal/httpapi/httpapi_test.go`) now sets `Events.MaxReplayEvents: 500`
  (production default) with a comment; before this ticket it stayed 0 and replay was
  silently skipped, which is why no test caught the leak.
- Tests: `internal/events/replay_test.go` (payload-shaped unit cases: drop/keep per
  step-scoped objectType, non-step-scoped pass-through, scope short-circuit, failure
  handling, stamp round-trip through `VisibleActivity`) and
  `internal/httpapi/blind_replay_test.go` (the repro inverted: blue stale-cursor replay
  asserts absence of the hidden step + execution ids and presence of `finding.created`;
  post-reveal replay of the same cursor delivers the events with `revealed: true`;
  red-unchanged; truncation ⇒ exactly one `stream.gap`; `MaxReplayEvents = 0` ⇒ no replay).
  Verified the regression fails against the pre-fix filter (reproduces the proof) and passes
  with the fix.

Pre-existing, unrelated failures on this tree (not introduced here): `make build`'s web
step fails on `src/api/query-client.test.ts` (`QueryClient.query` — clean tree too), and
`internal/config/types_test.go` fails gofmt on a clean tree.
