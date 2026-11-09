# Onion backend: Go to Rust migration (single PR)

Source of truth for the migration. `STATE.md` tracks progress, `PROMPT.md` drives the agent, `scripts/rust-capped.sh` caps its resource use.

## 1. Delivery model

One long-lived branch (`rust-migration`), one draft PR. The branch ports the whole backend to Rust, swaps the build and deploy pipeline, and deletes the Go code in its last commits. **Merging the PR is the cutover**: a merge to `main` tags and deploys changed targets (`onion-api`, `onion-worker`), so the first production run of the Rust backend happens at merge. The PR therefore stays a draft until every row in `STATE.md` is done, and a human runs the pre-merge gate (section 11).

Consequences of the single-PR model:

- Go keeps building and its tests keep running on the branch until Phase 6 deletes it. Go is the parity oracle.
- No strangler, no proxy middleware, no per-module ownership flags, no Go-to-Rust bridges. Coexistence only happens locally and in CI, never in production.
- One commit per unit (`feat(rust): <unit id> <what>`), so review and `git revert` work per unit. Push at most once per phase (plus fixes for failing CI): the review bot has a small daily quota.

## 2. Facts this plan rests on

Measured in the checkout:

- `services/onion`: ~470k lines of hand-written Go (non-test), ~175k lines of Go tests, 25 modules in `internal/modules`, ~44k lines in `internal/platform`, `internal/ai/classifier` (a model checkpoint plus Go inference, ~2.7 MB of weights), `internal/compat` (~2.6k lines). Entrypoints `cmd/api` and `cmd/worker`.
- Modular monolith with explicit wiring in `cmd/api/main_modules.go` / `main_routes.go`. `architecture_test.go` already scans import edges between modules (reuse it for the coupling graph).
- Contracts: 357 `.onk` files in `voxie/api`; `onek` pinned at **v0.36.3**, which supports `rust-server` and `rust-client`. Verified on a scratch copy of the contracts: 729 Rust files / ~155k lines, `cargo check` clean with axum 0.8, serde, serde_json, tokio, tower, http, bytes, futures, async-stream, base64, uuid, chrono, validator 0.20, regex, url. The Go and Rust output for the same contracts must agree.
- Generated server shape: per service `trait XService: Send + Sync + 'static` with `fn method(&self, RequestContext, Req) -> impl Future<Output = Result<Resp, XMethodServerError>> + Send`, and `x_service_router(Arc<T>) -> Router`. `RequestContext` carries headers, method, uri, extensions, `required_scopes`, `meta`.
- Not everything is contract-generated: ~49 `@http` raw routes, hand-written compat routes (`cmd/api/compat_routes.go`, `internal/compat`), `/health`, `/version`, the middleware chain in `internal/platform/server/http/web_server.go` (1,000 req/s per client key rate limit, Partials `X-API-KEY` bypass, origins), WebSocket bus with topic authorization and optional Redis backplane, SSE, and 9 `go:embed` uses (SQL schemas, email templates).
- Jobs coordinate through a MySQL locker (`platform/jobs/mysql_locker.go`) and the event bus is MySQL-backed (`platform/eventbus`). Rust must use the identical lock keys, job names, schedules, and outbox semantics, because during an ECS rolling deploy the old Go task and the new Rust task briefly run together.
- SQL migrations are plain SQL under `internal/modules/*/sql` and `internal/platform/*/sql`, run by `scripts/migrate.sh` (`Dockerfile.migrate` copies `scripts/` and `internal/`; `migrate-onion-database.yml` and `.github/scripts/migration-key.sh` read them). **These paths must not move.** sqlc output is generated and gitignored.
- `.github/deploy-targets.json` defines `onion-api` (image `voxire/onion`, container `api`, port 9999, `health_url` `/health`, `version_url` `/version`) and `onion-worker` (image `voxire/onion-worker`, container `worker`), both with a Go-specific `go_package` field consumed by `build-ecs-image.yml` and `.github/scripts/build-onion-binaries.sh`. `go-build-cache.yml` saves Go caches; `onion-images.yml` builds both images on pull requests without pushing or touching AWS.
- `media-engine` is ~5k lines of C++ over libvips and ffmpeg (ffmpeg is spawned as a subprocess). Go talks to it over a framed stdio protocol (`internal/platform/cdn/engine.go`, `--serve`).
- Worker health: ECS silently rolled back worker deploys when the worker did not serve on `PROBE_PORT`. The Rust worker must serve the same health endpoint on `PROBE_PORT`.

## 3. Goals and non-goals

Goals: no Go left in `services/onion`; identical HTTP behavior, identical environment variables, images, ports, health and version endpoints; contracts stay the single source (`voxie/api` to generated Rust).

Non-goals: new features, schema changes, redesign of module boundaries, frontend changes, upgrading `onek` (v0.40 replaced `.onk` with TypeSpec; separate project), touching `rtylrs/infra`, AWS, or any production system from the agent.

## 4. Layout

The Rust workspace root is `services/onion/` itself (`Cargo.toml` next to `go.mod`), so `services/onion/**` path filters, `release.yml`, and Dockerfiles keep working. Rust code does not live under `internal/`.

```
services/onion/
  Cargo.toml, rust-toolchain.toml, .cargo/config.toml    workspace, pinned stable, build caps
  crates/
    contracts/      committed Cargo.toml + lib.rs; src/generated/ is gitignored onek output
    core/           config, env, errors, money, pagination, cursor, csvsafe, ids, utils
    db/             sqlx MySQL pool, tx helpers, tenant scoping, sea-query helpers
    auth/           jwt, authorization, tenant, origins, securekey, emailcrypto
    events/         eventbus, jobs (MySQL locker + scheduler), emailqueue
    realtime/       wsbus + Redis backplane
    integrations/   s3/cdn, email (resend), slack, paddle, trino, sheets, openai, gemini
    ai/             classifier checkpoint loading and inference
    server/         middleware chain, health, version, error writer, embed helpers
    media-engine/   binary: same --serve protocol as the C++ engine
    modules/<name>/ one crate per Go module (25) plus compat, layered domain/application/infrastructure/jobs
  bins/api, bins/worker
  tools/            sqlc2sqlx, route-dump, parity
  internal/modules/*/sql, internal/platform/*/sql        stay exactly where they are
```

One crate per module gives parallel and incremental compiles. At Phase 6, delete `*.go`, `go.mod`, `go.sum`, `sqlc.yaml` files, Go-only scripts, and `architecture_test.go` (replace the latter with a Rust check). Keep SQL, `scripts/migrate.sh`, `Dockerfile.migrate`, and anything the migration pipeline reads.

## 5. Technology mapping

| Go | Rust |
| --- | --- |
| net/http + onek go-server | axum 0.8 + onek rust-server traits |
| sqlc + Squirrel | sqlx (MySQL, rustls, `query_file_as!`, offline `.sqlx` data) via `tools/sqlc2sqlx`; sea-query where Squirrel built dynamic SQL |
| SQL migrations | unchanged |
| goroutines, errgroup, sync | tokio tasks, JoinSet, tokio::sync |
| slog + betterstack, sentry-go | tracing + tracing-subscriber, sentry |
| go-redis | redis (tokio) |
| golang-jwt | jsonwebtoken |
| aws-sdk-go-v2 s3 | aws-sdk-s3 |
| coder/websocket | axum ws |
| robfig/cron | in-house scheduler matching `platform/jobs` (cron crate for parsing) |
| openai-go | async-openai |
| genai (Gemini), Trino, Google Sheets/Drive, Resend, Paddle | reqwest against the REST APIs, only endpoints actually used; yup-oauth2 for Google |
| ledongthuc/pdf | lopdf or pdf-extract (verify on fixtures) |
| fogleman/gg, x/image | tiny-skia, image, imageproc |
| go-nude | no equivalent: port from Go source with fixtures, or delete if unused |
| classifier (Go inference) | reimplement the forward pass in Rust; numeric parity against `testdata` and `real_weights_test.go` fixtures |
| go:embed | rust-embed or include_str!/include_bytes! |
| validator / protovalidate | generated `validator` code plus hand checks where Go used custom validation |

Rules: no global singletons (Go `wsbus.SetDefault` becomes a field on `AppState`); `AppState` holds `Arc`'d services built once in `bins/api`. `thiserror` enums per crate mapped to the generated per-method error enums at the handler edge. Money is a decimal type, never `f64`. No `unwrap`/`expect`/`panic` outside tests and startup. `clippy -D warnings`. Same environment variable names and semantics as Go (`.env.example`, `check_env_example.py`).

## 6. Module recipe

1. Contract: confirm the module's services generate; implement each `XService` on a `ModuleService` that delegates to application services.
2. Domain: port types and invariants 1:1 from `internal/domain`.
3. Infrastructure: run `tools/sqlc2sqlx` on the module's `sql/`; port repositories; same transaction boundaries and SQL semantics. Ownership columns mean different things per module (e.g. `hr_time_entries.owner_id` is the business ID): port exactly as Go does.
4. Application: port services. Cross-module dependencies become traits defined in the owner crate and injected through `AppState` (replaces Go `Set*` wiring).
5. Jobs and event handlers: identical names, schedules, lock keys, outbox semantics.
6. Registration: `pub fn register(state: &AppState) -> ModuleRegistration` returning routers, jobs, handlers.
7. Tests: port the module's Go tests with the same fixtures and assertions.
8. Parity: route table equals Go's for the module; HTTP replay diff is clean.

## 7. Pipeline swap (Phase 5, in this PR)

Keep every external name and number identical: image repositories, container names, port 9999, `/health`, `/version` payload (uses `VERSION`, `COMMIT`, `BUILD_TIME` build args), user `app`, all environment variables.

- `services/onion/Dockerfile`: build the Rust `api` and `media-engine`, runtime image keeps libvips and ffmpeg, the classifier model directory, and email templates; `FORM_MAPPING_CHECKPOINT_PATH` and `MEDIA_ENGINE_PATH` stay.
- `services/onion/Dockerfile.worker`: small image, serves health on `PROBE_PORT`.
- `services/onion/Dockerfile.migrate`: unchanged unless it breaks.
- `.github/deploy-targets.json`: replace `go_package` with the Rust equivalent; update `build-ecs-image.yml` and `.github/scripts/build-onion-binaries.sh` to match.
- `go-build-cache.yml` becomes a Rust cache workflow (Swatinem/rust-cache on free runners); `onion-images.yml` keeps building both images on PRs and is the gate for this phase.
- Do not build images locally. CI does it.

## 8. Verification

All required for a module to be marked done:

1. Route parity: `tools/route-dump` vs the Go route table (`route_dump_test.go`) including methods, scopes, `@meta`, compat and raw routes.
2. HTTP replay: `tools/parity` seeds a throwaway MySQL, runs Go and Rust against identical data, replays a corpus (success, validation errors, authorization failures, pagination, tenant isolation), and diffs normalized responses.
3. Ported tests with the same fixtures.
4. `scripts/check_schema_drift.py`, `scripts/check_env_example.py`, `scripts/check-route-contract-drift.py` still pass.

CI uses free GitHub-hosted runners and Actions cache only: `cargo fmt --check`, `clippy -D warnings`, `cargo check --workspace`, non-DB unit tests, and `onion-images.yml` image builds. DB-backed suites run locally against the throwaway MySQL (`scripts/test-db.sh`, container `onion-env-db-1`), same as Go today.

## 9. Resource caps (the agent runs on a small Linux server)

- Every heavy command (cargo, go, make, docker, test-db) runs through `services/onion/scripts/rust-capped.sh`. It enforces a hard memory cap and CPU quota with a systemd cgroup, refuses to start if disk is low, sets `CARGO_BUILD_JOBS`, `RUST_TEST_THREADS`, `GOMAXPROCS`, `-p` for Go, and runs at low priority. It fails closed if it cannot apply the cgroup.
- `.cargo/config.toml` sets `jobs`, `debug = 0` for dev and test profiles, and `incremental = false`.
- Prefer `cargo check -p <crate>` over workspace builds; never run `cargo build --release` locally. Never run `docker build` locally (CI does it). One heavy process tree at a time. No parallel agents.
- Cap the throwaway MySQL: `docker update --memory 1g --memory-swap 1g --cpus 1 onion-env-db-1`.
- If a command is killed (exit 137, OOM, timeout), shrink the scope (`-p`, `-j 1`, fewer tests) and retry; do not raise the caps. Raising caps is the human's call.

## 10. Phases

| Phase | Work | Gate |
| --- | --- | --- |
| 0 Foundations | Workspace + caps + CI gates; contracts crate from `onek` `rust-server`; non-contract inventory; coupling graph and port order; Go route dump and `tools/route-dump`; spike (one raw route, one service trait); media-engine port; parity harness skeleton | Workspace builds in CI; INVENTORY.md and COUPLING.md committed; route-dump works; media-engine golden tests pass |
| 1 Platform | `core`, `db`, `auth`, `events`, `realtime`, `integrations`, `ai`, `server`, config/env parity, `AppState` and bin skeletons with `/health`, `/version`, `PROBE_PORT` | Ported tests green; jobs and eventbus pass locker and outbox tests on throwaway MySQL |
| 2 Query layer | `tools/sqlc2sqlx` converts all query packages | Converted queries compile with offline data; schema drift check passes |
| 3 Modules | Port in the order from COUPLING.md; leaf modules first (webhooks, notifications, cacheops, emailtest, aiquota, aiquery, activity, hub_menu, websitelead, automations, integrations, insights, portal), then caret, booking, support, recruit, store, flow, crm, then foundational and heavy ones as the graph dictates (organization, finance, erp, pos, hr), plus `compat` | Per module: recipe steps 1-8 done, one commit per module |
| 4 Composition | `bins/api` registers every module; `bins/worker` runs every job and handler; full-app route parity and replay | Whole-app route table equals Go; replay diff clean |
| 5 Pipeline swap | Section 7 | `onion-images.yml` builds both Rust images in PR CI |
| 6 Delete Go | Remove Go code and Go tooling, update `AGENTS.md` and docs | CI green with Rust only; STATE all done; PR marked ready |

## 11. Pre-merge gate (human, not the agent)

Because merge deploys:

1. All PR checks green, including both image builds.
2. Run the PR's images against a restored copy of production data in a non-production environment and run `tools/parity` there. Review the auth, money, finance, and tenant-isolation commits by hand.
3. Confirm infra needs with `rtylrs/infra` (task definitions, memory and CPU, `PROBE_PORT`, env vars, health checks) and promote any task-definition change.
4. Rollback plan ready: redeploy the previous Go image tag (check `redeploy.yml` supports a tag). Merge in a low-traffic window.
5. Squash-merge (repo policy), then watch `/health`, `/version`, error rates, job runs, and worker health.

## 12. Risks

- Behavior drift in money, finance journals, tenant isolation, authorization. Mitigated by replay and by porting tests first, and by the pre-merge review, not by trust.
- A single atomic cutover has no gradual rollout. The pre-merge gate and rollback plan are the safety net.
- Rolling deploy overlap of Go and Rust tasks: identical lock keys and outbox semantics are mandatory.
- Non-contract surface (compat, raw routes, middleware, websockets) is easy to miss. INVENTORY.md exists to prevent that.
- Classifier numeric parity, PDF extraction parity, and the Gemini, Trino, and Google API clients are the weakest ports.
- A ~600k-line diff cannot be reviewed as a whole; commit-per-unit is the review unit.
- The 49 `@http` raw routes are unverified on the Rust generator until the spike.

## 13. Effort

Honest range: 12 to 24 months for one person with heavy AI assistance, bottlenecked on verification and review rather than code generation. Phases 0 to 2 take 2 to 4 months. A small server stretches wall-clock time because it forces `cargo check` per crate and serial work.
