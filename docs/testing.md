# Testing

Four layers, run by three commands. This page is mostly about the fourth, because it is the one
with moving parts.

| Layer | Command | What it drives |
|---|---|---|
| Go unit and store | `make test` / `make test-race` | Packages, and a temp DuckDB file for the store |
| Web unit | `make test` | `vitest` + Testing Library, with MSW standing in for the API |
| Container | `make docker-smoke` | The image: boots with no network, persists, has Chromium |
| **End to end** | **`make e2e`** | **A browser against the real binary and a real database** |

[`PLAN.md`](../PLAN.md) §9 sets the bar per layer. Coverage is reported, never gated.

## The end-to-end suite

[`e2e/`](../e2e) is a Playwright project of its own. It drives `bin/blacklight` — the binary, with
this build's SPA embedded — over HTTP, against a DuckDB file created for the run and deleted after
it. Nothing is mocked. If it passes, those pieces genuinely fit together.

```sh
make e2e-browsers   # once, and again after a Playwright upgrade (~150 MB)
make e2e            # builds the SPA and both binaries, then runs the suite
```

`make e2e` builds first on purpose: a suite run against yesterday's binary is worse than no suite,
because it is green about the wrong thing.

Flags go through `E2E_ARGS`, or run Playwright directly from `e2e/`:

```sh
make e2e E2E_ARGS='--grep version'
cd e2e && npx playwright test specs/smoke.spec.ts
```

### It fails when there is nothing to test

This is the point of the harness, and it is worth knowing before the first time you see it.

v1's `global-setup.ts` called `process.exit(0)` when nothing answered on `BASE_URL`. Every run was
green and none of them ran a test — which is worse than having no end-to-end tests at all, because
it buys confidence nothing earned. `PLAN.md` §9 names it.

The replacement has two modes and neither can pass silently:

- **`BASE_URL` unset** — the harness owns the servers. Each spec file gets its own `blacklight`
  process on its own fresh database. A server that never becomes healthy fails the run, quoting the
  tail of its own log.
- **`BASE_URL` set** — you are pointing the suite at a server you started. Global setup probes
  `$BASE_URL/api/v1/healthz` and **throws** if nothing healthy answers within the budget:

  ```
  Error: No healthy Blacklight server answered at http://127.0.0.1:45999/api/v1/healthz
    waited:       3065 ms of a 3000 ms budget
    last attempt: fetch failed: connect ECONNREFUSED 127.0.0.1:45999
  ```

  ```sh
  make run &                                 # in one terminal
  BASE_URL=http://localhost:8080 make e2e    # in another
  ```

### Every spec file starts from a clean database

Each spec file gets its own server and its own DuckDB file, and `--repeat-each` gets a fresh one per
repeat. So the suite means the same thing whatever order it runs in, and a single file run alone
behaves like the same file run last. Prove it whenever you have touched the harness:

```sh
cd e2e
npx playwright test --repeat-each=2
npx playwright test specs/smoke.spec.ts specs/isolation.spec.ts   # and the other way round
npx playwright test specs/isolation.spec.ts                       # each file alone
```

Playwright has no `--shuffle`; naming the files in different orders is how you vary it.

Tests **within** one file share that file's server and run in order, in one worker. That is where a
sequence like "create a scenario, then execute a step in it" belongs. Files run in parallel across
workers.

### Seeding

Put the system into a known state with `test.use({ seed: { steps: [...] } })` at the top of a spec
file. Each step is an argument vector for [`blctl`](../cmd/blctl), run in order against that file's
fresh database **before its server boots**:

```ts
test.use({
  seed: {
    steps: [
      ['user', 'create', '--email', 'red@example.test', '--role', 'red'],
      ['content', 'sync', '--source', 'attack'],
    ],
  },
})
```

A step that reads stdin takes the object form. `blctl user create` wants the password there and
deliberately not in a flag, because a flag ends up in shell history and in `ps`:

```ts
test.use({
  seed: {
    steps: [
      {
        args: ['user', 'create', '--email', 'lead@example.test', '--name', 'Lee', '--admin'],
        stdin: 'the password this spec signs in with',
      },
    ],
  },
})
```

The `{ steps: [...] }` wrapper is not decoration. Playwright reads any option value that is an array
whose second element is an object as a `[value, options]` fixture tuple, and a bare array of two
seed steps is exactly that — the option would silently resolve to the first step alone. `SeedPlan`
in `harness/server.ts` carries the full explanation.

Two reasons it is `blctl` and not SQL. A fixture full of `INSERT`s drifts away from what the
application actually writes, and the first thing it stops exercising is the validation that would
have caught the bug. And DuckDB admits one writer to a file at a time, so a seed step *during* a
test would fail on the running server's lock — before the server starts is not a limitation, it is
the only correct place.

A failing seed step fails the spec file loudly. It never runs the tests against a half-seeded
database.

`blctl` has the subcommand tree since `M0B-014` — `blctl --help` lists it, and
[`docs/cli.md`](cli.md) explains it. All seeded commands above (`user create`,
`content sync`, etc.) are implemented and tested. A spec file whose seed step
fails (bad args, missing source, locked database) fails loudly, which is the
correct outcome — the seed system itself is the assertion.

## Debugging a failure

Everything below runs from `e2e/`.

```sh
npx playwright test --headed          # watch it happen in a real window
npx playwright test --ui              # the UI mode runner: step, pick locators, re-run one test
npx playwright test --debug           # Playwright Inspector, paused at the first action
```

After a failed run, four things are waiting in `test-results/<test>/`:

| File | Answers |
|---|---|
| `trace.zip` | What the browser did, step by step, with a DOM snapshot at each one |
| `test-failed-1.png` | What the screen looked like when it gave up |
| `video.webm` | How it got there |
| `blacklight-server.log` | What the **server** thought of it — requests, status codes, errors |

The server log is the one that is usually missing from other people's suites, and the one that
usually has the answer. It is attached to the failing test automatically.

```sh
npx playwright show-trace test-results/<test>/trace.zip
make e2e-report     # the HTML report of the last run, with all of the above linked
```

When the question is what was actually *in* the database, keep it:

```sh
BLACKLIGHT_E2E_KEEP=1 npx playwright test
# → BLACKLIGHT_E2E_KEEP is set; databases and server logs kept in /tmp/blacklight-e2e-XXXX
```

Each spec file's directory under there holds its `blacklight.duckdb`, its `evidence/`, and its
`server.log`. Open the database with `duckdb`. Nothing else in the run refers to it, so delete it
when you are done.

### Writing specs

- **Query by role, label or text** — `getByRole('link', { name: 'Health' })`, not a class chain.
  The UI is Tailwind, so a class chain is a list of styling decisions: it breaks when someone
  adjusts a margin, and still passes when the value it was meant to check is wrong. Reach for
  `data-testid` only when no role or label can express the target.
- **No sleeps.** `page.waitForTimeout` asserts nothing and fails on a slow machine. Web-first
  assertions (`await expect(locator).toHaveText(...)`) already retry until they pass or time out.
  `eslint-plugin-playwright` fails the build on both this and an unconditionally skipped test.
- **`await` everything.** An un-awaited `expect` passes whatever the application did.
  `no-floating-promises` catches it.
- **No retries are configured**, deliberately. A retry turns a flaky test green and hands the cost
  to whoever debugs it in three months. Flakiness here is a bug — in the product or in the spec.

The suite grows one milestone at a time towards the single spec `PLAN.md` §9
describes: install content, run a baseline and retest engagement with
cross-engagement comparison, and share a report. M6 delivered the product
thesis (`M6-015`); the thesis spec exercises it end to end.

### Thesis spec (M6-015)

`specs/thesis-report.spec.ts` exercises the full product thesis:

| Step | What |
|---|---|
| Content | Seeded from offline ATT&CK + Atomic fixtures via `blctl` |
| Baseline engagement | Created via API: two steps (T1059, T1190), red executes (pending→running→complete), blue scores |
| Finding | Raised from the missed T1190 detection |
| Retest engagement | Second engagement (M5 rewrite — no rounds): T1190 re-run, scored higher |
| Report | Created and published via API |

The report is published through the API on the retest engagement. Share creation,
guest claim, HTML view, and revoke → 404 are exercised by
`specs/reports-share.spec.ts` (M6-012).

**Known pre-existing bugs discovered during M6-015 implementation** (all in
`internal/report/` — need separate fixes):

1. **DuckDB params scan** — `putReportBlocks` fails with "unsupported Scan,
   storing driver.Value type map[string]interface {} into type *string".
   Blocks cannot be set via API.
2. **DuckDB nil scan** — `createReportShare` fails scanning `<nil>` into
   `*json.RawMessage` for versions with no blocks. Blockless versions are
   un-shareable via API.
3. **Authz engagement mapping** — `putReportBlocks` maps `reportId` as the
   engagement resource identifier instead of `engagementId`. Admin workaround
   used; member sessions get 404.

## Workbook delete and reorder (known pre-existing bugs)

A coverage pass over `internal/engagement` found the workbook's scenario/step
delete and reorder paths were never exercised, and two of them do not work.
Both causes are DuckDB limitations `internal/store/engagement/engagements.go`
already documents for `deleteEngagementGraph`:

- DuckDB binds parameters for the **first statement of a multi-statement `Exec`
  only**, so a semicolon-separated script with a `?` in every statement fails
  with "incorrect argument count for command".
- A `RESTRICT` foreign key is enforced against the child's index, not the
  current transaction's own deletes, so a child and its parent cannot be removed
  in one transaction. DuckDB has no other `ON DELETE` action, and it implements
  `UPDATE` as delete+insert, so **updating a row that a RESTRICT FK references
  fails too**.

What is fixed here (see `internal/store/engagement/scenarios.go`,
`findings.go`, and the tests in `internal/engagement/workbook_test.go`):

- **`Scenarios.Delete`** — was one multi-statement script; now one committed
  statement at a time, the `deleteEngagementGraph` pattern. `DELETE` of a
  scenario (with or without children) works.
- **`Findings.Delete`** — deleted `finding_step` and the finding in one
  transaction, which the RESTRICT FK rejects; the child rows now commit before
  the parent. `DELETE` of a finding works.

Still broken, needing a design decision rather than a mechanical fix:

- **`Steps.Delete`** — same multi-statement script as `Scenarios.Delete`, plus a
  renumber step (`UPDATE app.step SET ordinal = ordinal - 1 …`). The renumber
  cannot succeed while the remaining steps have executions, because
  `app.execution.step_id` is a RESTRICT FK and DuckDB rewrites the `UPDATE` as
  delete+insert. Every step the service creates has an execution, so a step
  delete is a 500 in production.
- **`Steps.Reorder` / `Scenarios.Reorder`** — assign ordinals one `UPDATE` at a
  time. A reorder of rows that have children (steps under a scenario,
  executions under a step) hits the RESTRICT FK, and even without children the
  first move can collide with the `(scenario_id, ordinal)` unique constraint
  before the occupant has moved. Both endpoints fail for any non-trivial
  workbook.

`internal/engagement/workbook_test.go` covers the rejection paths (incomplete
set, foreign id, closed engagement) that do work; the success paths are omitted
until the storage approach changes.

## Content sync write fairness (M2-016)

Content Apply is the largest write volume in the system and shares the single
serialized DuckDB writer with interactive paths (session touch, user update).
`internal/content/loadtest` proves Apply batching does not starve those paths.

### CI (always on)

```sh
go test ./internal/content/loadtest/
```

| Test | What it proves |
|---|---|
| `TestSyncWriteFairness` | Multi-batch fixture sync + session-touch every 50ms; interactive **p95 ≤ 200ms**, max ≤ 2s, sync succeeds |
| `TestSyncWriteFairnessDetectsLockHold` | Same setup with Apply sleeping **250ms inside** `store.Write`; interactive p95 **exceeds** 200ms — the gate would catch a lock held across work |

Both ride along with `make test` / `make test-race`. A failure on the first means
shrink `BLACKLIGHT_CONTENT_WRITE_BATCH` or stop holding `store.Write` across
non-trivial work. A failure on the second means the detector itself regressed
(HoldWrite no longer inside the lock, or the budget went soft).

### Full developer load

```sh
BLACKLIGHT_LOADTEST=1 go test -count=1 -timeout 15m ./internal/content/loadtest/ -run TestSyncWriteLoad
```

Writes **20 000** fixture notes at the production default batch size (250) while
probing the same way. Pass/fail uses the same p95 budget. Ballpark on a local
SSD / arm64 devcontainer: ~10s sync, interactive p95 well under 200ms.

Assumptions: single process, local NVMe/SSD (or the codespace disk), no competing
writers. Not a multi-node test and not the M3 war-room scoring load.

ATT&CK checked-in fixtures are deliberately tiny (a handful of STIX objects);
the multi-batch volume under test is the synthetic fixture adapter, which is
what production adapters already share for `Writer` batching.

## War-room concurrency (M3-016)

Twenty concurrent users updating executions, uploading evidence, and posting
comments share the single serialized DuckDB writer. `internal/engagement/loadtest`
proves the optimistic-lock gate catches lost updates and interactive latency stays
within budget. **This is the gate before M4–M6.**

### CI (always on)

```sh
go test ./internal/engagement/loadtest/
```

| Test | What it proves |
|---|---|
| `TestWarRoomConcurrency` | 5 users × 20 steps, 5s: red/blue patches, evidence, comments, reads. Interactive **p95 ≤ 200ms**, max ≤ 2s, zero lost updates. |
| `TestWarRoomConcurrencyDetectsLostUpdates` | Same setup with version WHERE clause removed from patches. The consistency check detects lost updates (version sums don't add up) — proves the gate catches the bug. |

Both ride along with `make test` / `make test-race`. A failure on the first means
the serialized writer is overloaded — check for `store.Write` held across
non-trivial work. A failure on the second means the version-check detector itself
regressed.

### Full developer load

```sh
BLACKLIGHT_LOADTEST=1 go test -count=1 -timeout 15m ./internal/engagement/loadtest/ -run TestWarRoomLoad
```

Runs **20 users** across **50+ steps** for 15 seconds at production concurrency.
Pass/fail uses the same p95 budget. Ballpark on a local SSD / arm64 devcontainer:
~15s, interactive p95 well under 200ms.

Assumptions: single process, local NVMe/SSD, no competing writers. Not a
multi-node test and not an SSE fan-out stress (M4).

### Retry helper

`patchRedWithRetry` and `patchBlueWithRetry` in the test demonstrate the client
409 → re-GET → re-PATCH pattern under contention: on version conflict, re-read
the execution to get the current version and retry up to 5 times.


## SSE war-room load (M4-010)

Twenty concurrent SSE subscribers, catch-up replay, presence heartbeats,
and slow-client eviction share the in-process SSE hub.
`internal/events/loadtest` proves publish p95 stays under budget under
concurrent subscriber load, stalled subscribers are evicted without
blocking publishers, and goroutines do not leak. **This is the gate
before M5–M6.**

### CI (always on)

```sh
go test ./internal/events/loadtest/
```

| Test | What it proves |
|---|---|
| `TestSSEWarRoomConcurrency` | 5 users × 10 steps, 30 activity rows, 10 SSE subscribers (2 stalled), 10s: red PATCHes via RecordAlone, presence heartbeats, catch-up with Last-Event-ID. Publish **p95 ≤ 200ms**, max ≤ 2s, events received, stalled channels closed, goroutine delta ≤ 100. |

Rides along with `make test` / `make test-race`. A failure means the hub
publish path is blocking on slow clients — check `Hub.Publish` and the
subscriber channel eviction path.

### Full developer load

```sh
BLACKLIGHT_LOADTEST=1 go test -count=1 -timeout 15m ./internal/events/loadtest/ -run TestSSEWarRoomLoad
```

Runs **20 users** × 2 tabs (40 SSE subscribers), 250 activity rows,
25 steps for 20 seconds. Pass/fail uses the same p95 budget. Ballpark
on a local SSD / arm64 devcontainer: ~25s, publish p95 well under 200ms.

Assumptions: single process, local NVMe/SSD. Real HTTP SSE against a
test server so authz and catch-up replay are exercised end to end.
Not a multi-node test and not a browser fan-out at scale.

The `TestSSEWarRoomConcurrency` CI gate verifies that stalled subscribers
(never-read channels) are evicted and publish continues — a mutation in
the hub that silently removes eviction would cause the CI gate to hang
or exceed the goroutine budget.

### Reverse proxy note

SSE through buffering proxies breaks live UX. Deploy docs note
`proxy_buffering off` for `/api/v1/events` (M2-004). M4-010 verifies
through the compose/deploy path, not only Vite dev.



## Analytics query budget (M5-015)

Every analytics rollup (coverage, MTTD, burndown, compare, Navigator) and every
export path must stay responsive under concurrent write load — the same queries
that feed the dashboard (M5-013) and reports (M6). `internal/analytics/loadtest`
proves queries stay within budget and the archive path holds constant memory.
**This is the gate before M6.**

### CI (always on)

```sh
go test ./internal/analytics/loadtest/
```

| Test | What it proves |
|---|---|
| `TestAnalyticsQueryBudget` | 200 techniques, 5 scenarios × 10 steps, 50 findings, 3 concurrent writers, 10s: every rollup and endpoint measured. **All p95 ≤ 250ms**, max ≤ 1s, dashboard set ≤ 1s, write p95 ≤ 200ms. Archive memory delta ≤ 50 MiB. |
| `TestAnalyticsQueryBudgetDetectsRegression` | Same fixture, but TechniqueCoverage is replaced with a per-technique Go loop (an N+1 rewrite of the single-statement rollup, run 20× to simulate repeated calls). The broken query **exceeds** the 250ms budget — proves the gate catches a regression that text diffs would miss. |
| `TestArchiveExportMemory` | Archive export streams to a ZIP via `io.Pipe`; heap sampled before and after. Growth ≤ 50 MiB — the archive is streamed, not buffered. |

All three ride along with `make test` / `make test-race`. A budget failure on
the first means a rollup is too slow — fix the query, add an index, or fix a
join before the N+1 bites in M6's PDF render. A failure on the second means the
mutation detector itself regressed: the broken loop no longer exceeds budget.
A memory failure means the archive path is buffering where it should stream.

### Full developer load

```sh
BLACKLIGHT_LOADTEST=1 go test -count=1 -timeout 15m ./internal/analytics/loadtest/ -run TestAnalyticsQueryLoad
```

Runs **800 techniques**, **10 scenarios × 50 steps**, **200 findings** with
**1000 status-history rows**, **100 evidence blobs**, and **5 concurrent
writers** for 15 seconds. Pass/fail uses the same budgets. Ballpark on a local
SSD / arm64 devcontainer: ~20s, all rollups well under budget.

### When a budget fails

1. **Check the query.** Is it still one statement? A Go loop over `rows` for an
   aggregate is an N+1 regression — `M5-EPIC` bans it.
2. **Check the indexes.** `EXPLAIN` the failing query. DuckDB's default indexes
   are good for primary-key joins; a join on an unindexed column on a large
   table (e.g. `finding_status_history.changed_at`) will table-scan.
3. **Check the join shape.** A `CROSS JOIN` or a missing `WHERE` clause
   produces a cartesian product that grows with the fixture — rebuild the
   fixture at full scale to reproduce.
4. **Do not add a cache.** Materialized views, rollup tables, and in-process
   caches add staleness. `M5-EPIC` defers caching until this test proves it is
   needed, and the test hasn't done that.
5. **Do not raise the budget.** The budget reflects the M6 PDF render timeout.
   If a query cannot meet it, the query must change — not the budget.

Assumptions: single process, local NVMe/SSD, no competing writers. Not a
multi-node test and not the M6 Chromium render path (measured in M7-008).

## Report render budget (M7-008)

Every report block path — HTML render, publish (snapshot + store), PDF —
must stay responsive under realistic data and concurrent write load.
`internal/report/loadtest` proves budgets hold, and a mutation gate proves
the CI catches an N+1 regression inside a render.

### CI (always on)

```sh
go test ./internal/report/loadtest/
```

| Test | What it proves |
|---|---|
| `TestReportRenderBudget` | 200 techniques, 5 scenarios × 10 steps, 50 findings, 50 render iterations. HTML render **p95 ≤ 1s**, max ≤ 3s. |
| `TestReportRenderDetectsRegression` | Same as above but `TechniqueCoverage` replaced with 3× per-technique N+1 subqueries. Render **exceeds** 3s max budget — proves the gate catches a query regression. |
| `TestReportRenderWithConcurrentWrites` | 3 concurrent writers (execution updates) + HTML renders for 10s. HTML render p95 ≤ 1s, write p95 ≤ 200ms (M3-016 interactive budget). |

All three ride along with `make test` / `make test-race`.

### Full developer load

```sh
BLACKLIGHT_LOADTEST=1 go test -count=1 -timeout 15m ./internal/report/loadtest/ -run TestReportRenderLoad
```

Runs **800 techniques**, **10 scenarios × 50 steps**, **200 findings** with
**5 concurrent writers** for 15 seconds. Measures HTML render under write load,
publish path, and PDF smoke (Chromium).

### Budgets

| Path | Budget |
|---|---|
| HTML render p95 | ≤ 1s |
| HTML render max | ≤ 3s |
| Publish (render + snapshot + store) p95 | ≤ 2s |
| PDF render | ≤ 30s (documented; Chromium) |
| Interactive write p95 (under render load) | ≤ 200ms (M3-016 budget) |

### M7-008 re-run results (2026-08-10)

Existing gates re-run against current `main`:

| Gate | Original p95 | Re-run p95 | Budget | Status |
|---|---:|---:|---|---|
| M3-016 war-room writes | 16.7ms | 25.8ms | 200ms | ✅ |
| M4-010 SSE publish | 17.4ms | 11.5ms | 200ms | ✅ |
| M5-015 analytics queries | 11.2ms (Coverage) | 16.1ms (Coverage) | 250ms | ✅ |

No silent >2× regression detected. All budgets intact.

### When a budget fails

1. **Check the queries.** Every analytics-backed block calls the same rollup
   functions as M5-015. If a render budget fails but the query budget passes,
   the issue is in the HTML assembly or a block Render method.
2. **Check the publisher.** The publish path snapshots the full HTML into the
   versions table — a large report with evidence inline can be several MB.
3. **Do not hollow budgets.** The render budget reflects the PDF timeout.
   If it cannot be met, fix the render — not the budget.

## What CI runs

Every job in [`.github/workflows/ci.yml`](../.github/workflows/ci.yml), on every pull request; see
[`docs/contributing.md`](contributing.md) for the table and for branch protection. The `End-to-end`
job uploads `e2e/playwright-report` and `e2e/test-results` as the `e2e-report` artifact on every
run, passing or not.
