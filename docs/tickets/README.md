# Blacklight v2 — Ticket Backlog

Every ticket here derives from [`PLAN.md`](../../PLAN.md). If a ticket and `PLAN.md` disagree,
`PLAN.md` wins — raise it rather than guessing. **Exception:** locked epic decision tables (e.g.
[`M3-EPIC.md`](done/M3-EPIC.md) “no rounds”, [`M4-EPIC.md`](done/M4-EPIC.md) collaboration) are
explicit, reviewed deviations — follow the epic for that milestone and update `PLAN.md` when the
deviation is permanent.

> **Standing deviation — rounds.** `PLAN.md` §2 (`round` table, `(step_id, round_id)` grain), §5
> (round-over-round) and §9 steps 6–7 describe retest rounds. M3 dropped them and
> [`M5-EPIC.md`](M5-EPIC.md) locked the replacement — ad-hoc **cross-engagement compare**. Both
> milestones are decided; `PLAN.md` has not been rewritten to match. Read the epics for rounds,
> not the plan.

## How to read a ticket

Each ticket is one file, named `<ID>-<slug>.md`, and contains:

| Section | Meaning |
|---|---|
| **Why** | The user-visible or structural reason this exists. Read this before the scope. |
| **Scope** | Explicitly in, and explicitly out. Do not widen it — open a follow-up ticket instead. |
| **Files** | Where the code goes. Paths are suggestions with a strong prior; deviating needs a reason. |
| **Acceptance criteria** | Checkable statements. A reviewer runs down this list. |
| **Tests** | What must exist before this is reviewable. |
| **Depends on** | Tickets that must be merged first. |
| **Size** | S ≈ half a day, M ≈ 1–2 days, L ≈ 3–5 days, for someone new to the codebase. |

A finished ticket moves to [`done/`](done/) with its acceptance criteria ticked, and is marked ✅ in
the tables below. Where the implementation had to deviate from the ticket, the reason is appended to
the moved file under **Implementation notes** — read those before starting a ticket that depends
on it.

## Definition of done (applies to every ticket)

A ticket is done when **all** of the following are true. Tickets do not restate these.

1. `make lint test build` is green locally and in CI.
2. `make generate && git diff --exit-code` is clean — no stale generated code.
3. New behaviour has tests at the layer described in the ticket's **Tests** section.
4. No new `TODO`, commented-out code, or `panic()` on a reachable path.
5. No package-level database global. Repositories and services take dependencies via constructor
   arguments (`PLAN.md` §6).
6. Anything a future reader would find surprising is a comment; anything an operator needs is in
   `docs/`.
7. The PR description says what was tested and how, and links this ticket file.

## Conventions the whole backlog assumes

- **Module path** — `github.com/bryanster/blacklight` (unchanged from v1).
- **Spec-first.** `api/openapi.yaml` is edited *before* the Go or TS side. Hand-writing a handler
  signature or a fetch call that isn't in the spec is a review rejection.
- **Errors.** All API errors use the single problem shape defined in `M0B-007`. No ad-hoc JSON.
- **Time.** Store UTC, `TIMESTAMP` columns, serialize as RFC 3339. Never format a time for display
  on the server.
- **IDs.** UUIDv7 (`github.com/google/uuid` v1.6+ `uuid.NewV7`), stored as `TEXT`. Sortable by
  creation, so `ORDER BY id` is a stable tiebreaker.
- **Migrations are append-only.** Never edit a migration that has been merged; add a new one.
- **SQL stays portable ANSI.** No DuckDB-only syntax outside `internal/store/duckdb/`, so the
  escape hatch in `PLAN.md` §1 stays real.

## Milestones

| Milestone | State | Tickets |
|---|---|---|
| M0a — Clean slate | ✅ done on branch `v2` (see note below) | — |
| M0b — Foundations | ✅ done — 14/14 | [14 tickets](#m0b--foundations) |
| M1 — Identity & access | ✅ done — 18/18 | [18 tickets](#m1--identity--access) |
| M2 — Content | ✅ done — 16/16 | [16 tickets](#m2--content) · [`M2-EPIC.md`](done/M2-EPIC.md) |
| M3 — Core domain | ✅ done — 16/16 | [16 tickets](#m3--core-domain) · [`M3-EPIC.md`](done/M3-EPIC.md) |
| M4 — Collaboration | ✅ done — 10/10 | [10 tickets](#m4--collaboration) · [`M4-EPIC.md`](done/M4-EPIC.md) |
| M5 — Analytics | ✅ done — 15/15 | [15 tickets](#m5--analytics) · [`M5-EPIC.md`](done/M5-EPIC.md) |
| M6 — Reporting | ✅ done — 15/15 | [15 tickets](#m6--reporting) · [`M6-EPIC.md`](done/M6-EPIC.md) |
| M7 — Cutover | ✅ done — 14/14 | [14 tickets](#m7--cutover) · [`M7-EPIC.md`](done/M7-EPIC.md) |


> **M0a note:** the annotated tag **`v1-final`** exists at `c053fb741ba953bc8f2e151c05f966db813ec8fc`.
> Untracked `files/` and `.env` were never in git and are accepted lost (`M7-EPIC`).

---

## M0b — Foundations

Goal: an empty but *correct* skeleton. At the end of M0b, `docker compose up` serves a React page
that calls a generated, spec-validated API endpoint backed by DuckDB, and CI proves it.

Build in this order — the dependency chain is real.

| ID | Title | Size |
|---|---|---|
| [M0B-001](done/M0B-001-repo-skeleton-and-tooling.md) ✅ | Repo skeleton, Go module, Makefile, pinned tooling | M |
| [M0B-002](done/M0B-002-config.md) ✅ | Typed configuration from environment | S |
| [M0B-003](done/M0B-003-duckdb-store-and-serialized-writer.md) ✅ | DuckDB connection, pools, serialized writer | L |
| [M0B-004](done/M0B-004-migrator.md) ✅ | Embedded SQL migrator + `schema_migrations` | M |
| [M0B-005](done/M0B-005-openapi-and-server-codegen.md) ✅ | `api/openapi.yaml` + oapi-codegen strict server | M |
| [M0B-007](done/M0B-007-error-model.md) ✅ | One error/problem model, end to end | S |
| [M0B-006](done/M0B-006-http-server.md) ✅ | chi server, middleware chain, request validation, shutdown | M |
| [M0B-008](done/M0B-008-spa-scaffold.md) ✅ | React + Vite + TS + Tailwind + shadcn/ui scaffold | M |
| [M0B-009](done/M0B-009-typed-api-client.md) ✅ | Generated TS client + TanStack Query wiring | M |
| [M0B-010](done/M0B-010-embed-spa.md) ✅ | Embed `web/dist` in the binary, SPA fallback routing | S |
| [M0B-011](done/M0B-011-docker.md) ✅ | Dockerfile (CGO + Chromium) and compose | M |
| [M0B-012](done/M0B-012-ci.md) ✅ | CI: lint, test, build matrix, codegen-drift gate | M |
| [M0B-013](done/M0B-013-e2e-harness.md) ✅ | Playwright harness that fails loudly | M |
| [M0B-014](done/M0B-014-blctl-skeleton.md) ✅ | `blctl` admin CLI skeleton | S |

## M1 — Identity & access

Goal: nobody can do anything they shouldn't, and there is one place that decides. Every ticket
here traces to a named defect in `PLAN.md` §4 — the regression cases are the point.

| ID | Title | Size |
|---|---|---|
| [M1-001](done/M1-001-identity-schema.md) ✅ | Users, sessions, memberships schema | M |
| [M1-002](done/M1-002-password-hashing.md) ✅ | Argon2id hashing + password policy | S |
| [M1-003](done/M1-003-local-login-sessions.md) ✅ | Local login/logout, session cookies, rotation | L |
| [M1-004](done/M1-004-login-throttling.md) ✅ | Login throttling and lockout | M |
| [M1-005](done/M1-005-csrf.md) ✅ | CSRF double-submit for cookie sessions | M |
| [M1-006](done/M1-006-totp.md) ✅ | TOTP enrolment and verification | M |
| [M1-007](done/M1-007-recovery-codes.md) ✅ | MFA recovery codes | S |
| [M1-008](done/M1-008-mfa-enforcement.md) ✅ | Admin-enforced MFA (closes the skip-enrolment hole) | M |
| [M1-009](done/M1-009-oidc.md) ✅ | OIDC discovery login + group→role mapping | L |
| [M1-010](done/M1-010-saml.md) ✅ | SAML 2.0 service provider | L |
| [M1-011](done/M1-011-service-tokens.md) ✅ | Scoped service tokens, actually enforced | L |
| [M1-012](done/M1-012-authz-policy.md) ✅ | Central `authz.Can` policy engine | L |
| [M1-013](done/M1-013-authz-middleware.md) ✅ | One authorization middleware, zero handler checks | M |
| [M1-014](done/M1-014-permission-matrix-tests.md) ✅ | Full role × action × resource matrix tests | M |
| [M1-015](done/M1-015-activity-log.md) ✅ | Append-only activity log | M |
| [M1-016](done/M1-016-user-management-api.md) ✅ | Admin user management API | M |
| [M1-017](done/M1-017-auth-ui.md) ✅ | Login, MFA, account and admin UI | L |
| [M1-018](done/M1-018-admin-token-management.md) ✅ | Administrative service token management | S |

## M2 — Content

Goal: reference content is installable from the UI, not seeded as a 1 GB git clone. Sources are
enabled, synced, versioned and disabled; ATT&CK versions coexist; Atomic structure is preserved.
Decisions are locked in [`M2-EPIC.md`](done/M2-EPIC.md).

Build roughly in this order — the dependency chain is real.

| ID | Title | Size |
|---|---|---|
| [M2-001](done/M2-001-content-schema.md) ✅ | `content` schema: sources, versions, jobs, raw snapshots, seed rows | L |
| [M2-002](done/M2-002-source-registry-api.md) ✅ | Source registry API, enable/disable/delete, authz actions | M |
| [M2-003](done/M2-003-adapter-and-job-runner.md) ✅ | Adapter interface + global DB-backed job runner | L |
| [M2-004](done/M2-004-sse-hub.md) ✅ | Minimal shared SSE hub + sync progress | L |
| [M2-005](done/M2-005-bundle-upload-and-reprocess.md) ✅ | Offline bundle upload + reprocess-from-raw | M |
| [M2-006](done/M2-006-attack-adapter.md) ✅ | ATT&CK Enterprise adapter (multi-version) | L |
| [M2-007](done/M2-007-attack-version-pin-surface.md) ✅ | ATT&CK version catalog & pin surface | M |
| [M2-008](done/M2-008-atomic-adapter.md) ✅ | Atomic Red Team adapter | M |
| [M2-009](done/M2-009-sigma-adapter.md) ✅ | Sigma adapter (technique-mapped rules) | M |
| [M2-010](done/M2-010-ctid-adapter.md) ✅ | CTID emulation-plan catalog adapter | M |
| [M2-011](done/M2-011-custom-content-api.md) ✅ | Custom content API: templates, rules, notes | M |
| [M2-012](done/M2-012-v1-format-import.md) ✅ | Import v1 `testcases.json` + knowledgebase YAML | M |
| [M2-013](done/M2-013-library-browser-ui.md) ✅ | Content library browser UI | L |
| [M2-014](done/M2-014-sources-admin-ui.md) ✅ | Sources admin UI: sync, bundle, status, reprocess | L |
| [M2-015](done/M2-015-custom-and-import-ui.md) ✅ | Custom editor + v1 import UI | M |
| [M2-016](done/M2-016-sync-write-load-test.md) ✅ | Sync write load test (serialized writer fairness) | M |

## M3 — Core domain

Goal: Engagement → Scenario → Step → Execution, scored in ATT&CK Evaluations terms, with evidence,
comments, and findings. **No retest rounds in v1** — operators recreate the assessment
(`M3-EPIC` decisions; deliberate `PLAN.md` deviation). Decisions are locked in
[`M3-EPIC.md`](done/M3-EPIC.md).

Build roughly in this order — the dependency chain is real. **M3-016 is a gate before M4–M6.**

| ID | Title | Size |
|---|---|---|
| [M3-001](done/M3-001-domain-schema.md) ✅ | `app` domain schema: engagement, scenario, step, execution, finding, evidence, comment | L |
| [M3-002](done/M3-002-engagement-crud.md) ✅ | Engagement CRUD, status lifecycle, attack pin, mode | M |
| [M3-003](done/M3-003-membership-api.md) ✅ | Engagement membership management API | M |
| [M3-004](done/M3-004-scenarios.md) ✅ | Scenarios CRUD + reorder (`workbook.write`) | M |
| [M3-005](done/M3-005-steps.md) ✅ | Steps CRUD, copy-on-use, soft freeze, reveal | L |
| [M3-006](done/M3-006-executions-red.md) ✅ | Executions — red side PATCH + optimistic lock | M |
| [M3-007](done/M3-007-executions-blue.md) ✅ | Executions — blue side PATCH + optimistic lock | M |
| [M3-008](done/M3-008-scoring-domain.md) ✅ | Scoring domain: category ordinal, modifiers, derived outcome, MTTD | M |
| [M3-009](done/M3-009-evidence-store.md) ✅ | Evidence blob store + upload/download API | L |
| [M3-010](done/M3-010-comments.md) ✅ | Comments on executions + edit history | S |
| [M3-011](done/M3-011-findings.md) ✅ | Findings + `finding_step` join | M |
| [M3-012](done/M3-012-ctid-import.md) ✅ | CTID emulation plan → Scenario import | M |
| [M3-013](done/M3-013-atomic-to-step.md) ✅ | Atomic / procedure template → Step | M |
| [M3-014](done/M3-014-engagement-ui.md) ✅ | Engagement UI: board / workbook | L |
| [M3-015](done/M3-015-scoring-ui.md) ✅ | Scoring UI: 5-button scale + modifiers | M |
| [M3-016](done/M3-016-concurrency-load-test.md) ✅ | War-room concurrency load test (**gate before M4–M6**) | M |

## M4 — Collaboration

Goal: one shared war room — SSE live updates derived from the activity log, presence, live
comments and activity rail, blind mode correct on the wire, reconnect catch-up via
`Last-Event-ID`. Decisions are locked in [`M4-EPIC.md`](done/M4-EPIC.md).

Build roughly in this order — the dependency chain is real. **M4-010 is a gate before M5–M6.**

| ID | Title | Size |
|---|---|---|
| [M4-001](done/M4-001-engagement-sse-topics.md) ✅ | Engagement SSE topics + per-topic authz | M |
| [M4-002](done/M4-002-activity-event-fanout.md) ✅ | Activity → engagement event fan-out | M |
| [M4-003](done/M4-003-frontend-event-consumption.md) ✅ | Frontend event consumption + precise cache invalidation | M |
| [M4-004](done/M4-004-reconnect-catchup.md) ✅ | Reconnection + `Last-Event-ID` + blind delivery filter | L |
| [M4-005](done/M4-005-live-workbook.md) ✅ | Live workbook updates + 409 conflict toast | M |
| [M4-006](done/M4-006-presence.md) ✅ | Presence: heartbeat API, registry, SSE, UI | L |
| [M4-007](done/M4-007-comment-threads-ui.md) ✅ | Live comment threads + lightweight unread | M |
| [M4-008](done/M4-008-activity-rail-ui.md) ✅ | Engagement activity rail UI | M |
| [M4-009](done/M4-009-blind-mode-e2e.md) ✅ | Blind mode end-to-end (SSE + Playwright) | L |
| [M4-010](done/M4-010-sse-load-gate.md) ✅ | SSE war-room load test (**gate before M5–M6**) | M |

## M5 — Analytics

Goal: the workbook becomes the numbers a programme is judged on — coverage, detection distribution,
protection rate, MTTD, findings burndown, and the baseline-vs-retest delta — computed in **SQL, not
application loops** (`PLAN.md` §5). One source, two consumers: every number the dashboard shows is
the number an M6 report block prints, from the same query. Decisions are locked in
[`M5-EPIC.md`](M5-EPIC.md).

Build roughly in this order — the dependency chain is real. **M5-015 is a gate before M6.**

| ID | Title | Size |
|---|---|---|
| [M5-001](M5-001-analytics-query-layer.md) | Analytics query layer + seeded fixture | L |
| [M5-002](M5-002-blind-query-fence.md) | Query-layer blind fence for step reads (M3 debt) | S |
| [M5-003](M5-003-finding-status-history.md) | `finding_status_history` migration + write path | M |
| [M5-004](M5-004-coverage-rollups.md) | Coverage rollups: technique and tactic, dual denominator | M |
| [M5-005](M5-005-detection-distribution.md) | Detection-category distribution, protection rate, outcome mix | M |
| [M5-006](M5-006-mttd-analysis.md) | MTTD percentiles with detected/undetected counts | M |
| [M5-007](M5-007-findings-burndown.md) | Findings burndown | M |
| [M5-008](M5-008-cross-engagement-compare.md) | Cross-engagement compare rollup | L |
| [M5-009](M5-009-analytics-endpoints.md) | Analytics read endpoints + blind scoping + authz | L |
| [M5-010](M5-010-navigator-layer-export.md) | ATT&CK Navigator layer export | M |
| [M5-011](M5-011-json-csv-exports.md) | JSON and CSV exports | M |
| [M5-012](M5-012-engagement-archive-export.md) | Engagement archive export (versioned, round-tripped) | L |
| [M5-013](M5-013-dashboard-ui.md) | Dashboard UI: heatmap and scorecards | L |
| [M5-014](M5-014-compare-ui.md) | Cross-engagement compare UI | M |
| [M5-015](M5-015-analytics-query-budget.md) | Analytics query budget (**gate before M6**) | M |


## M6 — Reporting

Goal: section-picker report builder over a **block registry**, one HTML rendering path shared by
draft preview, immutable published versions, login-required share views, and PDF (`PLAN.md` §5, §8).
M6 is the usability bar — the branch is mergeable once it lands. Decisions are locked in
[`M6-EPIC.md`](M6-EPIC.md).

Build roughly in this order — the dependency chain is real. **M5-015 is a gate before M6.**
**M6-015** (PLAN.md §9 E2E thesis, M5 rewrite) is the M6 exit gate.

| ID | Title | Size |
|---|---|---|
| [M6-001](M6-001-block-registry.md) | Block registry (id, params schema, data deps, renderer hook) | M |
| [M6-002](M6-002-report-document-model.md) | Report document model, draft CRUD, `report.write` | L |
| [M6-003](M6-003-report-templates.md) | Engagement-scoped report templates | M |
| [M6-004](M6-004-branding.md) | Install branding defaults + per-report overrides | M |
| [M6-005](M6-005-rich-text-sanitization.md) | TipTap + server HTML allowlist (bluemonday) | M |
| [M6-006](M6-006-narrative-blocks.md) | Narrative blocks: cover, exec summary, scope/RoE, rich text, page break | M |
| [M6-007](done/M6-007-analytics-blocks.md) ✅ | Analytics blocks: heatmap, scorecard, distribution, gaps, MTTD, compare | L |
| [M6-008](M6-008-detail-blocks.md) | Detail blocks: scenario walkthrough, findings backlog, evidence appendix | L |
| [M6-009](M6-009-html-renderer.md) | Single HTML rendering path + golden files | L |
| [M6-010](M6-010-pdf-chromedp.md) | PDF via headless Chromium (`chromedp`) | M |
| [M6-011](M6-011-publish-and-versioning.md) | Publish, immutable versions, evidence opt-in, lead scope | L |
| [M6-012](done/M6-012-share-links.md) ✅ | Share links, grants/guests, password gate, revoke → 404 | L |
| [M6-013](M6-013-builder-ui.md) | Builder UI: blocks, reorder, params, HTML preview | L |
| [M6-014](M6-014-publish-share-ui.md) | Publish / versions / share & guest-grant UI | M |
| [M6-015](M6-015-e2e-thesis.md) | Complete PLAN.md §9 E2E thesis (M5 rewrite) | L |

## M7 — Cutover

Goal: ship **`v1.0.0`** — docs and README operator-ready, GHCR multi-arch + GitHub Release, backup /
upgrade path, security and performance checklists green, historical `v1-final` tag in place.
M6-015 remains the product thesis gate; M7 is the shippability gate. Decisions are locked in
[`M7-EPIC.md`](done/M7-EPIC.md).

Build roughly in this order — the dependency chain is real. **M7-001 is unblocked and should land
immediately.** **M7-009** is the exit gate. **M7-010…M7-012** are High findings from
[`docs/SECURITY_FINDINGS.md`](../SECURITY_FINDINGS.md) and are ship gates; **M7-013** and
**M7-014** are Medium and are not silent-defer.

| ID | Title | Size |
|---|---|---|
| [M7-001](done/M7-001-tag-v1-final.md) ✅ | Tag pre-rebuild tip as `v1-final` | S |
| [M7-002](done/M7-002-readme-rewrite.md) ✅ | README rewrite (product front door) | M |
| [M7-003](done/M7-003-docs-consolidation.md) ✅ | Docs consolidation & operator readiness | L |
| [M7-004](done/M7-004-release-workflow.md) ✅ | Release workflow: GHCR + GitHub Release + changelog | L |
| [M7-005](done/M7-005-upgrade-and-backup.md) ✅ | Upgrade path, backup procedure, optional `blctl backup` | M |
| [M7-006](done/M7-006-cutover-hygiene.md) ✅ | Cutover hygiene: v1 refs, CI branches, status banners | S |
| [M7-007](done/M7-007-security-review.md) ✅ | Security review pass (checklist) | L |
| [M7-008](done/M7-008-performance-sanity.md) ✅ | Performance sanity pass (re-run gates + report load) | M |
| [M7-009](done/M7-009-ship-v1.md) ✅ | Ship gate: `v1.0.0` release | M |
| [M7-010](done/M7-010-engagement-list-authz.md) ✅ | Engagement list leaks every engagement (BL-001, High) | S |
| [M7-011](done/M7-011-ownership-facts-loader.md) ✅ | Replace production Ownership.Facts stub (BL-003, High) | L |
| [M7-012](done/M7-012-cross-engagement-idor.md) ✅ | Bind nested IDs to the authorized engagement (BL-002, High) | L |
| [M7-013](done/M7-013-share-token-logging.md) ✅ | Share tokens in logs; unthrottled claim/password (BL-005, Medium) | M |
| [M7-014](done/M7-014-content-sync-ssrf.md) ✅ | Content sync SSRF allowlist (BL-004, Medium) | M |

## Security findings (BL)

Findings from security passes over the shipped tree. BL-001…BL-005 came out of the M7 review and
are closed by the `M7-010…M7-014` tickets above. Later findings are filed here directly, each
with its proof.

| ID | Title | Severity | State |
|---|---|---|---|
| [BL-006](done/BL-006-sse-replay-blind-leak.md) | SSE catch-up replay delivers unrevealed-step events to blue | Medium | ✅ closed 2026-09-09 |
| [BL-007](BL-007-finding-read-blind-step-leak.md) | Findings list exposes unrevealed step ids to blue | Medium | open |
