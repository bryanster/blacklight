# BL-007 — Findings list exposes unrevealed step ids to the blue seat

**Type:** security finding (blind-mode information disclosure) · **Severity:** Medium
**Depends on:** — · **Size:** S · **Status:** open · **Proven:** in-tree
(`internal/httpapi/blind_finding_step_leak_test.go`, passing)

## Why

`finding.read` carries no `GuardBlindMode` (`internal/authz/policy.go:271-273` — the guard applies
only to `execution.read`, `execution.write_blue` and `evidence.read`), and `findingToWire`
(`internal/httpapi/findinghandlers.go:166-197`) returns every linked step id with no blind
filtering. In a blind engagement, blue calling `GET /engagements/{id}/findings` or
`GET /findings/{fid}` receives the **ids of unrevealed steps** in the `stepIds` array — the exact
fact blind mode withholds ("not even to learn it exists", [`docs/authz.md`](../authz.md)).

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

- [ ] Blue in a blind engagement calling `GET /engagements/{id}/findings` receives findings with
      **no** `stepIds` entry for an unrevealed step.
- [ ] `GET /findings/{fid}` behaves the same, for every route that renders a finding
      (list, get, create, patch, steps endpoints).
- [ ] Revealed step links are still returned; red/lead/admin output unchanged.
- [ ] `TestBlindFindingStepIdsLeakUnrevealedSteps` asserts absence and passes.

## Tests

One regression at the handler layer (the existing demonstrating test, inverted), one case proving
revealed links survive the filter. Drive the HTTP API, not the struct.
