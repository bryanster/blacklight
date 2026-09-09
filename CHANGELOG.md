# Changelog

All notable changes to Blacklight are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **The SSE catch-up replay path now honours blind mode (BL-006, Medium).**
  Reconnecting with `Last-Event-ID` replays catch-up events rebuilt from the
  activity log, and those rebuilt payloads carried no `revealed` field — so
  the blind-mode filter that protects the live stream failed open on the
  replay path, and a blue member reconnecting with a stale
  `Last-Event-ID` received `step.*`, `execution.*`, `comment.*` and
  `evidence.*` frames about **unrevealed steps** while the live stream would
  have dropped the identical events. Replay now re-derives reveal state from
  the engagement store at delivery time: rows for unrevealed steps are
  dropped, events for revealed steps still arrive (including a step revealed
  between event write and replay), red/lead/admin replay is unchanged, and
  `MaxReplayEvents = 0` still disables replay with `stream.gap` on
  truncation.

## [1.0.3] — 2026-09-05

### Added

- **Client SDKs for Go, TypeScript/JavaScript, Python and Rust**, in
  [`sdk/`](sdk/). All four are generated from `api/openapi.yaml` — the same
  document the server is generated from — so a script written against any of
  them is checked against the API by its own compiler rather than by trial and
  error against a running deployment. `make generate` rewrites all four and the
  `codegen` job in CI fails if a spec change was not regenerated, which means an
  endpoint added, renamed or removed reaches every SDK in the same commit as the
  server. Each has one small hand-written seam holding the two things the
  document cannot say: that a deployment's origin needs `/api/v1` appended (the
  document's one server is relative, because the SPA is same-origin), and how a
  service token is presented. Each uses the best generator its language has
  rather than one tool for all four — the document is OpenAPI 3.1 and uses it,
  and `openapi-generator` reads 3.1 by converting it to 3.0, where "absent" and
  "explicitly null" stop being different things. Rust is the exception and the
  reason a container image is pinned at all: no Rust generator reads 3.1. See
  [`docs/sdk.md`](docs/sdk.md).
- **`make test-sdk`** — the four SDKs' own tests, and proof that each still
  compiles. A new `SDK tests` job runs it in CI.
- **A red-team guide with screenshots of every workflow**
  ([`docs/red-team-guide.md`](docs/red-team-guide.md)) and worked Python SDK
  examples that drive those workflows end to end
  ([`sdk/python/examples`](sdk/python/examples)): `red_team_agent.py` signs in,
  runs a scenario workbook, and files findings; `blue_team_agent.py` is the
  defender half. Both use only the generated client.
- **A NixOS deployment example** ([`deploy/nixos-azure/`](deploy/nixos-azure/)):
  a flake that deploys the container to an Azure VM, TLS included, with one
  `deploy.sh` invocation.
- **An Azure ARM deployment template**
  ([`deploy/azure-arm/`](deploy/azure-arm/)): the container image plus the
  PostgreSQL and storage sidecars, with session and encryption secrets written
  to Key Vault. The `terraform` example already covered the same ground; this
  is the template route for deployments that are not already on Terraform.

### Changed

- **The `ImportPlanResponse` schema is now `ImportPlanResult`.** The Go client
  generates a wrapper type per operation named `<OperationId>Response`, in the
  same package as the schema types, so a schema with that name alongside the
  `importPlan` operation declared one name twice and `sdk/go` did not compile.
  `…Result` is what the document already uses for this shape (`LoginResult`,
  `GuestRegisterResult`). The wire format is unchanged — the field names and the
  JSON a caller sends and receives are identical — but the generated type's name
  changed in the SPA's types and in the server's. A new rule in
  `api/conventions_test.go` refuses the collision from now on.
- **The Go toolchain floor is now 1.26, and the dependency tree is refreshed
  across Go, the web client, the e2e suite, the GitHub Actions and the
  Terraform example.** `chromedp v0.16` needs Go ≥ 1.26, so the pinned toolchain
  moves with it — `deploy/Dockerfile` and the devcontainer image build on
  `golang:1.26`, and golangci-lint is pinned to `v2.13.2`. The lint findings
  that newer release adds are cleared: report and export rendering use
  `fmt.Fprintf` instead of `WriteString(fmt.Sprintf(…))`, and the deliberate
  cookie attribute choices (configurable `Secure` for plain-http dev hosts, the
  script-readable CSRF cookie, the SAML flow's `SameSite=None`) are documented
  where they are made. The refresh closes the open security advisories:
  `@tiptap/core` is past GHSA-cp6q-959q-f8rh, and `fast-uri` is past the four
  high-severity advisories (a web development dependency).

### Fixed

- **A spoofed `Authorization: Bearer` header no longer moves a login or
  report-share throttle count off the account being guessed.** The bearer
  credential was checked before the route's, and `tokenAccount` maps a
  malformed token to the empty string, which the account limiter ignores — so a
  caller could attach any bearer header and escape the per-account lockout
  entirely. Sign-in and share-claim routes now name their credential from the
  body, cookie or URL path first, and the bearer token is only the credential
  in view when no route has named another.

## [1.0.2] — 2026-08-15

### Added

- **"Use in scenario" works, from the library side.** The button on a technique,
  procedure template, and emulation plan had been a disabled placeholder since
  the library shipped: importing needs an engagement to import _into_, and the
  library has none in scope. It now asks which one. Only engagements that would
  accept the content are offered — closed and archived ones refuse new scenarios
  server-side, so they are not listed — and choosing one hands off to that
  engagement's Workbook, where the import dialog opens with the object already
  chosen. Nothing is imported from the library itself: the confirmation stays on
  the engagement, which is where the ATT&CK pin and the workbook role apply. A
  read-only role or a closed engagement gets told so instead of a dialog it
  could not have used. The hand-off travels as `?use=&useId=` on the workbook
  URL and is stripped once consumed, so a reload does not re-open the dialog.
- **A first-run setup wizard, and it is a MITRE ATT&CK version picker.** A fresh
  installation boots with an empty content library — nothing is fetched at first
  boot — so the first administrator to sign in used to land in a product that
  could not map a step to a technique, with no prompt to say so. They now land
  on `/setup`, which lists every Enterprise release MITRE publishes, newest
  first, marks anything already installed, and installs the chosen one: enabling
  the seeded ATT&CK source and starting the same sync job the sources screen
  starts. Members are not redirected — an empty library is not theirs to fill.
  On a host with no route to MITRE the screen says so, carries the transport
  error, points at the offline bundle path, and still lets a release be
  installed by label. Skipping finishes setup without installing anything, which
  is the right answer for an air-gapped deployment. See
  [`docs/first-run-setup.md`](docs/first-run-setup.md).
- **`GET /content/attack/releases`** — the ATT&CK releases upstream offers,
  merged with what is installed, in MITRE's own order. An unreachable index is a
  `200` saying so rather than a `502`: air-gapped installations are supported,
  and for them that is the normal case.
- **`blctl setup status` / `complete` / `reset`.** `complete` marks first-run
  setup done without installing anything, for a provisioning run that has
  already created an administrator and synced content — and for the end-to-end
  suite. `reset` brings the wizard back and is deliberately not an endpoint.
- **A stopwatch on the red side of the step view.** MTTD is
  `detected_at − started_at`, and until now the red half of that subtraction had
  no control in the interface at all: a start time only appeared if the server
  happened to stamp one on a `→ running` transition, so the measurement blue's
  Detected At fed into was either an accident of when somebody changed a
  dropdown or missing entirely. The red panel now opens with a clock. **Start**
  records the moment it is pressed — it writes immediately rather than waiting
  for Save, because a stopwatch that saves later measures the save — moves a
  pending execution to Running, and the clock ticks until **Stop**, which
  records the end and completes the execution. **Started At** and **Ended At**
  sit under it for a run that happened away from the workbook or needs
  correcting afterwards, and an end before its start is refused at the field
  rather than by a `400`. The resulting MTTD is shown in the bar under the step
  title, between the two panels, since neither team owns it: a duration once
  both times exist, and which half is missing when they do not.

### Changed

- **The step view puts red and blue side by side.** Opening a step in the
  workbook stacked the red execution above the blue detection in a dialog barely
  wider than a phone, so the two halves of a purple team exercise were never on
  screen together and reconciling an execution against a detection meant
  scrolling between them. They are now two columns — red left, blue right, each
  under its own coloured header, stacking again below `md` — in a dialog wide
  enough for both, with the step's status, derived outcome and Reveal to Blue
  gathered into one bar under the title instead of taking a row each. The panel
  colours are theme tokens, so they hold up in dark mode rather than being two
  hardcoded reds.

### Fixed

- **Deleting a report from the reports list works, and it took two fixes.** The
  server end was the DuckDB foreign-key limitation this codebase has hit before:
  a RESTRICT constraint is checked against the child's index, which does not see
  the deleting transaction's own work, so removing `report_block` and `report`
  in one transaction failed with "still referenced by a foreign key" — on any
  report that had a block, which is any report anybody had opened. Deletion now
  walks the whole graph one committed statement at a time: share grants, shares,
  published versions, draft blocks, then the report. The browser end was
  separate and would have outlived the first fix: the mutation wrapped a 204
  endpoint in `unwrap`, which throws when there is no body, so a delete the
  server had honoured came back as "unexpected 204 response", left the confirm
  dialog sitting open, and never refreshed the list. It uses `unwrapVoid` now.
  The confirmation copy no longer promises that published versions survive —
  `report_version` has a RESTRICT key to `report`, so they never could.
- **The reports list shows how many blocks each report holds.** It said "0
  blocks" for every row, always. The row measured `report.blocks`, an array the
  list response has never carried — the spec says so in the field's own
  description — so it was counting an absent value. `Report` now carries a
  `blockCount` the list fills from a single grouped query, and the row reads
  that. The count is on every report response, including the one a rename
  returns, so a page that re-reads the list after an edit cannot flip back to
  zero.
- **The analytics MTTD panel reports an engagement that is still running.** It
  measured only executions red had concluded, so a step that had been started
  and detected — the ordinary state of an exercise in progress, and now the
  normal one, since the step view's Start button leaves an execution `running`
  until Stop — was dropped from the percentiles, from the detected count and
  from the denominator alike. The panel said "nothing scored yet" for a step
  whose own step view was showing an MTTD. MTTD now measures every execution
  red has begun (`complete`, `blocked` or `running`), which is one status wider
  than the engagement-wide attempted definition every other rollup uses, and
  deliberately so: MTTD is detection latency, not attempt coverage, and the
  latency exists from the moment red starts. `pending` and `skipped` are still
  excluded — neither has a start to measure from. Expect an engagement mid-run
  to report MTTD while Coverage and Protection rate still show nothing;
  [`docs/analytics.md`](docs/analytics.md) § MTTD carries the reasoning. The
  empty panel now says no executions have been _started_, which is what it has
  always meant, rather than blaming blue for not scoring.
- **Detection and execution times keep their seconds.** The step view's
  timestamp fields were minute-resolution, which rounded every MTTD by up to 59
  seconds and could round a detection back before the start it was measured
  from — a `400` on save for a detection that really did follow the attack. An
  unedited timestamp is now also left out of its PATCH rather than echoed back
  through the field, so an unrelated save no longer drops the milliseconds off a
  time the server recorded.
- **Moving an engagement through its lifecycle works.** Activate, Close and
  Archive answered 500 and left the engagement where it was, on every
  engagement — creating one seats its lead, and one referencing row was all it
  took. `app.engagement(status)` was indexed; DuckDB rewrites an `UPDATE` that
  touches an indexed column into `DELETE` + `INSERT`, and the delete half is
  checked against the `RESTRICT` foreign keys that members, scenarios, findings
  and reports hold the engagement down with. The index is gone — status has four
  values over a table with tens of rows, so it was never earning its keep — and
  the schema now says why, because the same trap is waiting for any index on a
  mutable column of a table other rows point at. The overview page also hid the
  transition buttons on a closed engagement, so even once the server answered,
  Archive was reachable only from Settings; archiving is the way _out_ of
  closed, and the page offers it again.
- **Importing a step from a template works again.** From Template in an
  engagement workbook hung for thirty seconds and then answered 500, every time
  — the step was never created. The activity entry recorded alongside the new
  step opened a second write transaction from inside the one creating it, and
  the store serializes writers: the recording queued behind the transaction that
  was waiting on it until the request deadline fired, and the commit then failed
  on a transaction the cancelled context had already rolled back. Recording now
  happens on the caller's transaction, as the hook was designed to, so the step
  and its activity row share one commit. Patching a step and patching a scenario
  recorded the same broken way and were failing identically; both are fixed with
  it.
- **The CTID content pack syncs again.** Its source is a GitHub archive of the
  whole `adversary_emulation_library` repository, which passed 512 MiB and so
  failed every sync with `download exceeds content max bytes limit`. The
  adapter no longer downloads it: it reads the repository listing, fetches only
  the plan files it actually parses, and reassembles them into a small zip with
  the archive's layout — about 75 KB moved instead of 640 MB, in a few seconds.
  Mirrors and offline bundles are unaffected, and the raw snapshot's digest now
  changes when a plan changes rather than on every upstream commit. See
  `docs/content-ctid.md`.
- **Administrators can edit engagements again.** An administrator opening an
  engagement got no workbook toolbar — no Add Scenario, Add Step, Import CTID or
  From Template — read-only red and blue execution editors, and no way to raise
  a finding. The server had always allowed all of it (every engagement-scoped
  rule grants `Platform: admins`); only the interface disagreed, so the buttons
  were absent rather than refused. The interface now reads the caller's actual
  seat first, falling back to the platform seat only for an administrator who
  holds no membership, and its permission predicates match the server's table
  for every role. This bit hardest on a fresh install, where the bootstrapped
  first account is an administrator and there is nobody else to be a lead yet.
- **Deleting an engagement works.** Delete engagement confirmed the deletion,
  said the engagement was gone and returned to the list — where it was still
  sitting. It had never worked, for three reasons at once: a rename in the
  migration that gave `engagement_member` its foreign key left DuckDB unable to
  delete any engagement row at all; the statement emptying the workbook was a
  multi-statement script with bound parameters, which DuckDB rejects outright;
  and it never covered the engagement's reports or report templates, both of
  which hold it down with a foreign key. Deleting now removes the whole graph —
  published report versions, share links and grants, finding status history, and
  the references evidence files hold on their blobs, so those become
  collectable. The interface no longer reports success for a delete the server
  refused. The delete is not atomic, because DuckDB will not let one transaction
  remove a row and then the row it references; one interrupted partway leaves
  the engagement partly emptied and still listed, and running it again finishes
  the job.
- **A new step now shows up in the workbook without a reload.** Creating a step
  refreshed only the per-scenario step list, which nothing on the workbook
  reads: the page renders from the whole-engagement step list and the executions
  list. So the step stayed invisible, and once a reload made it appear its
  drawer had no execution behind it — no status, no red or blue editor, no
  comments or evidence to attach. Both lists, and coverage, are now refreshed by
  every step mutation and by the matching live event, so a step arrives complete
  for the person who created it and for everyone watching.
- **Add Step and Add Step From Template open with a scenario already chosen.**
  The toolbar picked one before opening the dialog, but the dialog read that
  choice once when the page mounted — before anything had been picked — so
  **Create** stayed disabled until the scenario was selected a second time by
  hand.

## [1.0.1] — 2026-08-14

### Added

- **First administrator from the environment.** `BLACKLIGHT_BOOTSTRAP_ADMIN_EMAIL`,
  `BLACKLIGHT_BOOTSTRAP_ADMIN_NAME` and one of
  `BLACKLIGHT_BOOTSTRAP_ADMIN_PASSWORD_FILE` / `BLACKLIGHT_BOOTSTRAP_ADMIN_PASSWORD`
  create that account at startup — but only on a database with no accounts at
  all, so the configuration is inert on every start afterwards and can never
  change an account that exists. It is for deployments where `blctl user create`
  means stopping the server to get at the database (Container Apps, Cloud Run,
  ECS); keep using the CLI everywhere else.
- **Azure Terraform example creates the first administrator.** Set `admin_email`
  and the configuration generates a password, stores it in Key Vault write-only
  beside the other two secrets, and passes it to the app;
  `terraform output admin_password_command` prints the command that reads it
  back.

## [1.0.0] — 2026-08-14

> **Greenfield only.** Blacklight v1.0.0 is a ground-up rebuild of the prior
> Python/Mongo application. There is **no migration path** from the old stack —
> a new installation starts with an empty database. The annotated tag `v1-final`
> exists on the pre-rebuild tree for anyone who needs the historical source.
> Back up the v1 deployment independently if its data is still needed.

### Added

- **Container image** as the supported deployment artifact — multi-arch
  (`linux/amd64`, `linux/arm64`), Debian-based, Chromium included for PDF
  rendering. Pull from `ghcr.io/bryanster/blacklight`.
- **Identity & access:** local accounts with Argon2id password hashing, TOTP
  multi-factor authentication with recovery codes, admin-enforced MFA, OIDC and
  SAML 2.0 single sign-on, scoped service tokens, CSRF protection, login
  throttling, and a central authorization policy engine with full role×action
  matrix tests.
- **Content library:** ATT&CK Enterprise (multi-version), Atomic Red Team,
  Sigma rules, and CTID emulation plans — synced from upstream or uploaded as
  offline bundles, with version pinning and a custom content API.
- **Core domain:** Engagements → Scenarios → Steps → Executions, scored in
  ATT&CK Evaluations terms (detection categories, modifiers, MTTD), with
  evidence uploads, threaded comments, and findings tracking.
- **Collaboration:** real-time war room over SSE with per-topic authorization,
  presence, live workbook updates, comment threads, activity rail, and blind
  mode for red/blue separation.
- **Analytics:** coverage rollups, detection distribution, protection rate,
  MTTD percentiles, findings burndown, cross-engagement comparison — all
  computed in SQL, served through a single query layer consumed by both the
  dashboard and report blocks.
- **Reporting:** block-registry report builder with narrative blocks, analytics
  blocks, detail blocks, rich-text editing, branded HTML rendering, PDF via
  headless Chromium, immutable published versions, and share links with optional
  password gates.
- **Admin CLI** (`blctl`) for migrations, user creation, content sync, and
  database inspection.
- **End-to-end test suite** (Playwright) that drives the real binary against a
  real database.

### Upgrade notes

1. **Read the release notes** for the version you are moving to. Breaking
   changes, new required environment variables, and migration counts are listed
   there.
2. **[Back up](../docs/deploy.md#backup-and-restore) your deployment.** Migrations
   are forward-only; the only path back to the previous release is restore from
   backup.
3. **Stop the server.** `docker compose stop` — DuckDB admits one process per
   file.
4. **Pull the new image.** `docker compose pull` (or `docker compose build` if
   building from source).
5. **Start.** `docker compose up -d`. Migrations run at startup.
6. **Verify.** Open `$BLACKLIGHT_BASE_URL` and sign in.

If the new release does not start, restore the backup and use the previous
image tag. A database that has been migrated forward cannot be opened by an
older binary.

### Changed

- The entire application was rebuilt from scratch in Go + React + TypeScript +
  DuckDB. The prior Python/Flask/MongoDB codebase is retired; its final state
  is tagged `v1-final`.

### Removed

- **MongoDB.** Blacklight stores everything in a single DuckDB file. No
  separate database process, no connection strings.
- **Retest rounds.** v1's round-over-round workflow is replaced by
  cross-engagement comparison. Operators recreate assessments rather than
  re-running steps within a locked structure.
- **Static asset directory.** The binary is the whole application; nothing is
  served from disk.

[1.0.3]: https://github.com/bryanster/blacklight/releases/tag/v1.0.3
[1.0.2]: https://github.com/bryanster/blacklight/releases/tag/v1.0.2
[1.0.1]: https://github.com/bryanster/blacklight/releases/tag/v1.0.1
[1.0.0]: https://github.com/bryanster/blacklight/releases/tag/v1.0.0
