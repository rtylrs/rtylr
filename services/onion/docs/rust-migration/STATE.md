# Migration state

The agent reads this first and updates it in the same commit as each unit. One row, one commit. Never mark a row `done` without its gate in `ARCHITECTURE.md` section 10 met.

Status values: `todo`, `in-progress`, `blocked`, `done`. Put commit hashes and blockers in Notes.

## Phase 0 Foundations

| Unit | Status | Notes |
| --- | --- | --- |
| 0.1 Cargo workspace, toolchain pin, `.cargo/config.toml` caps, CI gates (fmt, clippy, check) on free runners | done | Workspace at services/onion (core only), 1.97.1 pin, jobs=2, debug=0, incremental=false; new .github/workflows/rust-check.yml (PR-only, no deploy) |
| 0.2 `contracts` crate fed by onek `rust-server` (output gitignored, wired into `make generate`) | done | contracts crate + rust-server in onekit.toml, copy to src/generated (gitignored) in voxie/Makefile; 729 files, rust-check generates before clippy/check |
| 0.3 `INVENTORY.md`: every non-contract surface (compat routes, `@http` raw routes, middleware, health/version, websockets, SSE, `go:embed`, env vars, jobs, event handlers); each item becomes a row below | done | INVENTORY.md (compat 9 pkgs/2631 LOC, 43 @http methods, middleware order, health/version/probe, wsbus, SSE, 9 embeds, 121 env keys, jobs/eventbus, 156 sql dirs) |
| 0.4 `COUPLING.md`: module coupling graph (reuse `architecture_test.go` edge scan), strongly connected components, port order | done | 33 edges, SCCs {hr,organization} and {erp,pos}, full 25-module port order in COUPLING.md |
| 0.5 Go route dump export and `tools/route-dump` | done | Go export kept via ROUTE_DUMP test (243 legacy rows, re-generated locally); tools/route-dump normalizes/sorts/diffs dumps (3 tests pass) |
| 0.6 Spike: one `@http` raw route and one service trait implementation end to end | done | tools/spike: ActivityService::list_activity (stub empty page) + CRM unsubscribe GET/POST /crm/public/unsubscribe 1:1 (no delivery/DB); 3 in-process tests pass |
| 0.7 media-engine Rust port with golden tests and benchmark | done | crates/media-engine: --serve framing codec + vips-free primitives (Size/Box, FitMode, Gravity, Color, numeric), 5 golden tests vs C++ *_test.cpp, perf example |
| 0.7b media-engine operations (`resize`, `transform`, `info`, `analyze`, `video`, `thumbnail` over libvips + ffmpeg subprocess) | todo | Phase 5 pipeline swap cannot drop the C++ build until these land; primitives-only coverage in 0.7 is not a full port |
| 0.8 `tools/parity` skeleton (throwaway MySQL, run Go and Rust, replay, diff) | done | tools/parity: JSONL corpus/run parse, canonicalize, diff (4 tests); diff clean exit 0, replay validates corpus; live seed/replay wires in Phases 3-4 |
| 0.9 `tools/arch-check`: port every rule in `architecture_test.go` (ports and adapters layering, module edges, table and migration ownership), reuse `docs/architecture-exceptions.json`, port the scanner self-test cases as fixtures, run in CI | done | tools/arch-check: line scanner over .rs (platform-import, domain-import, transport-tag, application-adapter-import, read-authority, foreign-persistence-import, generated-public-api, cross-module-transaction, module-cycle), migration-owner via schema_registry.tsv, sweep table/query checks carried as reported Go findings (15 `foreign-table-access`: organization writing into caret/flow/hr/notifications/pos tables — Go parity, reported non-gating), 13 tests pass (8 scanner + tags/platform/cycle/tx extras + CSV); CI runs `arch-check --root .` |

## Phase 1 Platform crates

| Unit | Status | Notes |
| --- | --- | --- |
| 1.1 `core` and config/env parity | done | onion-core: full AppConfig + env/getters, money (cents, Go JSON shape), csvsafe, pagination, ids/utils; 15 tests pass, check_env_example green |
| 1.2 `db` | done | crates/db: sqlx pool (25/10/30m/10m, 10s acq), DSN metadata + safety guards ported, tx helpers, TenantScope (explicit resolver), business scope conditions; 11 tests pass |
| 1.3 `auth` (jwt, authorization, tenant, origins, securekey, emailcrypto) | done | crates/auth: HS256 JWT (iss+exp, expired vs invalid), EffectiveAccess predicates, tenant re-export, origin allowlist, securekey, AES-GCM emailcrypto; 18 tests pass |
| 1.4 `events` (eventbus, jobs, emailqueue) | done | crates/events: encode/dispatcher (fan-out before mark), registry/scheduler with MySQL lock naming, email queue with fenced claim SQL, encryption wiring; 15 tests pass |
| 1.5 `realtime` | done | crates/realtime: wsbus hub + authorizer cache TTL, topic scopes, fan-out with lazy unsubscribe, 256-drop publish cap, flattened wsbus-events backplane; 6 tests pass |
| 1.6 `integrations` | done | crates/integrations: S3 (aws-sdk-s3, LoadDefaultConfig creds, Floci/path-style defaults, eu-west-1), CDN sign (HMAC path+"\n"+canonical query, 1h/30d), Resend (Retry-After → RateLimited 429/60s, 100ms outbound pacing), Slack chunk/batcher, Paddle (quantity 1, `Paddle-Idempotency-Key`, ts:body HMAC ±5m), Trino (TLS scheme, header auth, nextUri bound 10k, 120s timeout), Sheets (svc-acct RS256 → oauth → `Leads!A:J` USER_ENTERED/INSERT_ROWS), AI (opencode→gemini→deepseek→openai + AI_PRIMARY_PROVIDER front, Gemini native generateContent models 3.8/3.7); 10 tests pass |
| 1.7 `ai` (classifier) | done | crates/ai (onion-ai): checkpoint safetensors parse (F32 only, size bounds, Python-json canonical + SHA-256 state-signature check, kind() codes), tinyx forward pass (embed 257×w, encoder + option-encoder, score head, collate masks incl. safe first-token), planner (lexical fallback, threshold/clamp, submit demotion, extraction/normalization); 13 tests pass incl. golden parity (params 14720, log diff <1e-4 vs testdata/forms golden, batch + pickle/empty rejection) |
| 1.8 `server` (middleware chain, rate limit, Partials key bypass, health, version, error writer) | done | crates/server: chain order (platformMiddlewares → bins/api wiring in 1.9), sliding-window limiter + per-route auth/MFA/short-secret budgets (1000/s global, Retry-After 429s), trusted-proxy client key (onion-core is_trusted_proxy_ip, CF-Connecting-IP honored only from trusted peers, XFF/X-Real-IP ignored), X-API-KEY constant-time bypass predicate, CORS allowlist (21 headers, Vary: Origin, preflight 200), body limit (4 MiB, resume/CDN exemptions) + 30s/5s timeouts with long-lived matcher, security-headers table + Server/X-Powered-By strip, sanitized error writer (SecureErrorResponse, recover writes body unless Upgrade), /version shape (dev/unknown/rustc 1.97.1); 14 tests pass |
| 1.9 `AppState`, `bins/api` and `bins/worker` skeletons, `PROBE_PORT` | todo | |

## Phase 2 Query layer

| Unit | Status | Notes |
| --- | --- | --- |
| 2.1 `tools/sqlc2sqlx` on one module | todo | |
| 2.2 Convert all query packages | todo | |

## Phase 3 Modules (order provisional until 0.4 lands)

| Module | Go LOC (non-test) | Status | Notes |
| --- | --- | --- | --- |
| webhooks | ~2.6k | todo | |
| notifications | ~2.2k | todo | |
| cacheops | | todo | |
| emailtest | | todo | |
| aiquota | | todo | |
| aiquery | | todo | |
| activity | | todo | |
| hub_menu | | todo | |
| websitelead | | todo | |
| automations | | todo | |
| integrations | | todo | |
| insights | ~7.5k | todo | |
| portal | | todo | |
| caret | ~23k | todo | |
| booking | ~14k | todo | |
| support | ~15k | todo | |
| recruit | ~16k | todo | |
| store | ~9k | todo | |
| flow | ~30k | todo | |
| crm | ~46k | todo | |
| organization | ~46k | todo | |
| finance | ~47k | todo | |
| erp | ~55k | todo | |
| pos | ~70k | todo | |
| hr | ~97k | todo | |
| compat (`internal/compat`, `cmd/api/compat_routes.go`) | ~2.6k | todo | |

## Phase 4 Composition

| Unit | Status | Notes |
| --- | --- | --- |
| 4.1 `bins/api` registers every module and non-contract route from INVENTORY.md | todo | |
| 4.2 `bins/worker` runs every job and event handler | todo | |
| 4.3 Whole-app route parity and replay diff clean | todo | |

## Phase 5 Pipeline swap

| Unit | Status | Notes |
| --- | --- | --- |
| 5.1 `Dockerfile` and `Dockerfile.worker` for Rust (names, port, health, env unchanged) | todo | |
| 5.2 `deploy-targets.json`, `build-ecs-image.yml`, `build-onion-binaries.sh`, Rust cache workflow, `onion-images.yml` | todo | |

## Phase 6 Delete Go

| Unit | Status | Notes |
| --- | --- | --- |
| 6.1 Delete Go code, `go.mod`/`go.sum`, sqlc configs, Go-only scripts and workflows | todo | |
| 6.2 Delete `architecture_test.go` (arch-check already enforces); update `AGENTS.md`, `CLAUDE.md`, docs | todo | |
| 6.3 Final verification pass, mark PR ready for review | todo | |

## Human action required

(Append infra, secrets, approvals, and pre-merge gate items here, with the commit that needs them.)
