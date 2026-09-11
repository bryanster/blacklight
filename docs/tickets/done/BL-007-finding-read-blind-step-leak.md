# BL-007 — Findings list exposes unrevealed step ids to the blue seat

**Type:** security finding (blind-mode information disclosure) · **Severity:** Medium
**Depends on:** — · **Size:** S · **Status:** closed (2026-09-11) · **Proven:** in-tree
(`internal/httpapi/blind_finding_step_leak_test.go`, passing — flipped to the
fixed behaviour and verified to fail against the pre-fix tree)

## Why

`finding.read` carries no `GuardBlindMode` (`internal/authz/policy.go:271-273` — the guard applies
only to `execution.read`, `execution.write_blue` and `evidence.read`), and `findingToWire`
(`internal/httpapi/findinghandlers.go:166-197`) returns every linked step id with no blind
filtering. In a blind engagement, blue calling `GET /engagements/{id}/findings` or
`GET /findings/{fid}` receives the **ids of unrevealed steps** in the `stepIds` array — the exact
fact blind mode withholds ("not even to learn it exists", [`docs/authz.md`](../../authz.md)).

The codebase already treats these links as blind-withheld one path over: the archive export
collects "findings — always all findings, but blind-filter linked step IDs"
(`internal/httpapi/archivehandler.go:118`). The API path is the exception. The committed test
below proves the leak end to end and passes today, so the disclosure is live, not theoretical:

```
$ GOTOOLCHAIN=auto go test ./internal/httpapi -run TestBlindFindingStepIdsLeakUnrevealedSteps
ok      github.com/bryanster/blacklight/internal/httpapi
```

(`blind_finding_step_leak_test.go:20-70`: blue's step GET is 404, blue's findings list returns the
same step id in `stepIds`.)

With a step id, blue confirms the step's existence and can correlate hidden structure across
findings (which hidden steps share a gap, severities, timing). Reading the step body stays 404 —
this is an existence/linkage leak, not a read bypass.

## Sink

| Where | What |
|---|---|
| `internal/authz/policy.go:271-273` | `finding.read` rule — no `GuardBlindMode` (contrast `:247,253,256`) |
| `internal/httpapi/findinghandlers.go:188-197` (`findingToWire`) | loads `FindingSteps` and emits every linked step id, unrevealed included |
| Contrast, correct: `internal/httpapi/archivehandler.go:118` | archive export blind-filters the same finding→step links |

## Scope

**In**

- Strip unrevealed step ids from the wire in `findingToWire` (list, get, create/patch responses),
  reusing the existing reveal-resolution helper (`resolveActivityStepID` /
  `evidenceConcealed`-style lookup) so the check is live at read time. Semantics match the
  archive path: **all findings remain visible to blue; only unrevealed `stepIds` links are
  withheld.** A finding whose links are all hidden simply arrives with an empty `stepIds`.
- Regression: convert the demonstrating test into the fixed behaviour — same seed, assert the
  findings list returns no unrevealed step id and still returns revealed ones; keep the 404
  assertion on the step itself.

**Out**

- Whether finding *bodies* should be withheld in blind mode (a finding title/description may
  reference hidden activity in prose) — that is a product decision, not this fix; the documented
  model says `finding.read` is membership-gated only. Raise separately if wanted.
- Moving `finding.read` under `GuardBlindMode` (would 404 whole findings for blue) — heavier than
  the defect and changes what blue may read; noted as the alternative, not chosen.
- BL-006 (SSE replay leak) — separate ticket, same family.

## Files

- `internal/httpapi/findinghandlers.go` — filter `stepIds` in `findingToWire`.
- `internal/httpapi/blind_finding_step_leak_test.go` — flip to the fixed behaviour (or move to
  the service layer if the filter lands there).
- `internal/httpapi/findinghandlers_test.go` — revealed-links-still-returned case.

## Acceptance criteria

- [x] Blue in a blind engagement calling `GET /engagements/{id}/findings` receives findings with
      **no** `stepIds` entry for an unrevealed step.
- [x] `GET /findings/{fid}` behaves the same, for every route that renders a finding
      (list, get, create, patch, steps endpoints).
- [x] Revealed step links are still returned; red/lead/admin output unchanged.
- [x] `TestBlindFindingStepIdsLeakUnrevealedSteps` asserts absence and passes.

## Tests

One regression at the handler layer (the existing demonstrating test, inverted), one case proving
revealed links survive the filter. Drive the HTTP API, not the struct.

## Implementation notes

Landed 2026-09-11.

- The filter landed as `(h *handlers) visibleFindingStepIDs` next to `findingToWire`, which is
  the single point behind list, get, create and patch (the steps endpoint answers 204 and renders
  nothing). Scope comes from the shared `stepBlindScope`; reveal state comes from `IsStepRevealed`
  — live at read time, so a step revealed after the finding was linked is returned and an
  unrevealed one is dropped, exactly the archive path's semantics (`archivehandler.go`).
- Error semantics differ from the event-delivery paths on purpose: a lookup failure fails the
  request (500) rather than keeping the link. The activity list and replay filter fail open
  because silently dropping a row loses history that cannot be re-derived; a read endpoint has
  nothing to preserve, so it matches the handler-side blind checks (`evidenceConcealed`), which
  propagate.
- The regression kept its demonstrating name (`TestBlindFindingStepIdsLeakUnrevealedSteps`, per
  the acceptance criterion) and now asserts absence: the same seed grew a revealed step linked to
  the same finding, the list and single-finding responses must contain only the revealed id
  alongside the 404/200 step assertions, and red's list must still carry both. Verified the test
  fails against the pre-fix filter and passes with it.
- `internal/httpapi/findinghandlers_test.go` is new (the ticket suggested it exist): pins that a
  standard engagement's blue seat receives the complete link set — including an unrevealed step,
  which a standard engagement does not hide — so the filter cannot over-apply, and that an
  unlinked finding renders `stepIds: []`.
- The patch leg of the regression is not driven over HTTP: `findings.Update`
  (`internal/store/engagement/findings.go`) currently fails with a DuckDB foreign-key violation
  for **every** seat when the finding has linked steps (`Violates foreign key constraint because
  key "finding_id: …" is still referenced`), pre-existing and unrelated to the filter. A patch
  response renders through the same `findingToWire`; worth its own ticket.
- Environment notes, not code: `make generate`'s Python SDK step needs `uv`, absent here (the
  aborted run's half-written output was restored from git; every other generated tree diffs
  clean); `golangci-lint` needs `GOLANGCI_LINT_CACHE` pointed somewhere writable;
  `internal/config/types_test.go:138` fails gofmt on a clean tree (as recorded on BL-006).
