# Non-contract inventory

Everything the Rust backend must reproduce that is not generated from
`voxie/api/*.onk`. Counts measured on `rust-migration` at Unit 0.3.
Composition (Phase 4, unit `4.1`) registers every item below; route parity
(`tools/route-dump` vs Go) covers each one.

## 1. Legacy compat routes

- Entry: `services/onion/cmd/api/compat_routes.go` (66 lines) registers 9
  generated packages from `services/onion/internal/compat/*/server.gen.go`
  (2,631 lines total): `crm`, `finance`, `hr`, `hr shift`, `organization`,
  `pos`, `recruit`, `recruit careers`, `store`.
- `internal/compat/<x>compat/server.gen.go` per package (69-780 lines each)
  serve the legacy route table via `main_routes.go`.
- These are hand-owned adapters, not onek output: reinstall port (`compat`
  module) with identical paths, query/body binding and `JsonResponse`
  (`data: json @unwrap`) pass-through.
- Rust: `crates/modules/compat` + registration in `bins/api`.

## 2. `@http` raw routes (43 methods)

Decoded from `voxie/api/**/service.onk` (comment mentions excluded).
Handler writes its own response; request/response `EmptyRequest`/`EmptyResponse`
are placeholders.

| Service (service.onk) | Methods |
| --- | --- |
| `recruit/careers/compat/v1` | `createApply` + `@max_body("16MiB")` |
| `recruit/compat/v1` (CompatibilityService) | `listPublicOg`, `createPublicApply`, `createPublicResumeUploadUrl` |
| `recruit/compat/v1` (VoxhireService) | 27 admin/apply paths (see file) |
| `pos/compat/v1` | `listWs`, `createPaymentsWebhook`, `createWsTicket` |
| `crm/compat/v1` | `getUnsubscribe`, `confirmUnsubscribe` |
| `organization/compat/v1` | `listResources`, `getResources`, `getResourcesRows`, `getResourcesRows2`, `listAuthorize`, `createPaddleWebhook` |
| `hr/compat/v1` | `createReportsGenerate` |
| `finance/compat/v1` | `@http` JsonResponse/BodyRequest compat routes (see file) |
| `store/compat/v1` | none (JsonResponse compat only) |

Per-route rate limits and admin session guards live in the original handlers
(comments in `recruit/compat/v1/service.onk`). Rust: port handlers 1:1.

## 3. Middleware chain and server

Source: `internal/platform/server/http/web_server.go`,
`register.go`, `response_compress.go`, `body_limit.go`, `request_logger.go`,
`auth_middleware.go`, `subscription_middleware.go`,
`partials_key` bypass, `observability` recover.

- `platformMiddlewares()` order (conditional `safeRequestLogger` first):
  recover, Sentry recover, CORS (origins), crash-notifier, security-headers,
  secure-error, response-compress, body-limit (`DefaultMaxRequestBodyBytes`),
  global rate limit (1000/s), `AuthRouteRateLimit`, request-timeout (15s),
  metrics counter, Flow-v1 usage report. Then run-mode `MiddlewareProvider`
  middlewares (auth, subscription enforcement, etc.).
- Partials `X-API-KEY` bypass wraps outside the composed handler
  (`PartialsBypassHandler`); attaches console-admin actor when valid.
- Server timeouts: `Read 30s`, `Write 30s`, `ReadHeader 10s`, `Idle 120s`,
  `MaxHeaderBytes 1MB`; shutdown graceful window 20s.
- Rust: `crates/server` (middleware chain, rate limit, bypass, error writer).

## 4. Health, version, probe, metrics

- `GET /health` in `web_server.go`: `{"status":"ok"}` or `503` when DB ping
  fails (2s timeout) on DB-backed run modes.
- `GET /version` in `version_routes.go`: `version`, `commit`, `build_time`
  (ldflags `Version`/`Commit`/`BuildTime`), `go_version` (`runtime.Version()`).
- `GET /{$}` and `GET /favicon.ico` registered in `web_server.go`.
- `GET /metrics` via `httpmetrics.Handler()` (basic-auth gated, opt-in).
- Worker/main probe server (`internal/platform/server/probe/probe.go`):
  `/health` (`{"status":"ok","service":...}`), `/ready`, `/stop` (token),
  `/metrics` (opt-in). `PROBE_PORT` disables probe when empty.
- External names unchanged: port 9999, `/health`, `/version`, env shape.

## 5. Real-time: websockets and SSE

- `internal/platform/wsbus` (topic auth in `cmd/api/main_routes.go`:
  `user:`, `business:`, `workspace:`, `project:`, `task:`; default deny) with
  optional Redis backplane.
- Handlers: `notifications/.../ws_handler.go`,
  `pos/orders/.../ws_handler.go`, `flow/workspaces/.../realtime_handler.go`
  (+ realtime tests).
- SSE: `caret/.../sse_handler_exceeded.go` (`Content-Type:
  text/event-stream`), `transcript.go`, `sse_handler_usage.go`,
  `background_reply.go`; `query_timeout.go` skips these from timeout.
- Rust: `crates/realtime` (axum ws + redis backplane), SSE handlers in
  `crates/modules/caret`.

## 6. `go:embed` assets (9 sites)

| File | Assets |
| --- | --- |
| `portal/internal/catalog/catalog.go` | `navigation.json`, `pages.json` |
| `caret/.../system_prompt.go`, `buddy_agent.go`, `countries.go`, `subagents.go` | `system_prompt.txt`, `assistant_prompt.txt`, `buddy_prompt.txt`, `subagent_prompt.txt`, `countries_data.json` |
| `internal/platform/email/render.go` | `templates/base.html`, `templates/templates/*/*.html` |

Rust: `include_str!`/`include_bytes!` or `rust-embed`, byte-identical.

## 7. Configuration and environment

- `.env.example` (255 lines, 121 keys) is the operator index; enforce with
  `scripts/check_env_example.py`. Runtime: `APP_ENV`, `RUN_MODE`, `PORT`,
  `SERVICE_NAME`, `TRUSTED_PROXY_CIDRS`; DB (`DATABASE_URL` or `DB_*`),
  auth (`JWT_SECRET`, `JWT_DURATION`, `PARTIALS_API_KEY`,
  `PARTIALS_ACTOR_USER_ID`, `PROBE_PORT`, `PROBE_STOP_TOKEN`); cache
  (`REDIS_*`), Trino (`TRINO_*`), S3/R2 (`S3_BUCKET_NAME`, `S3_REGION`,
  `S3_ENDPOINT`, `S3_FORCE_PATH_STYLE`), plus provider keys (OpenAI/Gemini/
  OpenCode/DeepSeek), Paddle, Google OAuth, SearXNG/Brave.
- Loader: `internal/platform/config/` (`config.go`, `config_part2.go`,
  `config_part3.go`, `appenv.go`, `dotenv.go`, `cache.go`, `paddle.go`,
  `trino.go`, `trusted_proxy.go`). ALLOW_REMOTE_DB and dev-vs-RDS guards in
  `config`/`appenv.go` port 1:1 (never touch a real DB).
- Rust: `crates/core` config + `crates/db` pool strings; identical names.

## 8. Jobs, event bus, email queue

- `internal/platform/jobs/registry.go` (`Registry.Register(Job)`),
  `scheduler.go` (robfig/cron v3 parsing, UTC, per-job `Name`/`Schedule`),
  `mysql_locker.go` (SQL lock keys), `autoregister.go`.
- `internal/platform/eventbus/` MySQL-outbox: `publisher.go` (`Record`),
  `dispatcher.go` (`NewDispatcher`), `jobs.go`, `reader.go`. Lock keys,
  job names, schedules and outbox semantics must match Go (rolling deploy
  runs old Go + new Rust together).
- `internal/platform/emailqueue/` (`queue.go`, `pool`,crypto/dedupe/lease/
  retention) wired in `cmd/worker/main.go` via `emailqueue.NewPool`.
- `cmd/worker/main.go`: `AutoRegister`, `eventbus.RegisterJobs`, scheduler
  start/stop, shutdown barrier. `PROBE_PORT` health for ECS.
- Rust: `crates/events` + worker job registration parity (Phase 4, `4.2`).

## 9. Data, contracts, and sidecars

- SQL: 156 dirs under `internal/modules/*/sql` and `internal/platform/*/sql`
  (plain SQL, applied by `scripts/migrate.sh` / `Dockerfile.migrate`;
  paths must not move; sqlc output is generated and gitignored).
- `media-engine/` (~5k C++ over libvips/ffmpeg) speaks framed stdio
  `--serve` (`internal/platform/cdn/engine.go`). Rust: `crates/media-engine`,
  same protocol.
- `internal/ai/classifier/` (checkpoint + Go inference, safetensors weights,
  `testdata/` golden fixtures). Rust: `crates/ai`, numeric parity.
- Contracts: 357 `.onk` files, 187 services, ~1,050 operations
  (`@get 427`, `@post 324`, `@put 165`, `@delete 100`, `@patch 34`).
