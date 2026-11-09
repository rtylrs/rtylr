# Migration state

The agent reads this first and updates it in the same commit as each unit. One row, one commit. Never mark a row `done` without its gate in `ARCHITECTURE.md` section 10 met.

Status values: `todo`, `in-progress`, `blocked`, `done`. Put commit hashes and blockers in Notes.

## Phase 0 Foundations

| Unit | Status | Notes |
| --- | --- | --- |
| 0.1 Cargo workspace, toolchain pin, `.cargo/config.toml` caps, CI gates (fmt, clippy, check) on free runners | done | Workspace at services/onion (core only), 1.97.1 pin, jobs=2, debug=0, incremental=false; new .github/workflows/rust-check.yml (PR-only, no deploy) |
| 0.2 `contracts` crate fed by onek `rust-server` (output gitignored, wired into `make generate`) | done | contracts crate + rust-server in onekit.toml, copy to src/generated (gitignored) in voxie/Makefile; 729 files, rust-check generates before clippy/check |
| 0.3 `INVENTORY.md`: every non-contract surface (compat routes, `@http` raw routes, middleware, health/version, websockets, SSE, `go:embed`, env vars, jobs, event handlers); each item becomes a row below | done | INVENTORY.md (compat 9 pkgs/2631 LOC, 43 @http methods, middleware order, health/version/probe, wsbus, SSE, 9 embeds, 121 env keys, jobs/eventbus, 156 sql dirs) |
| 0.4 `COUPLING.md`: module coupling graph (reuse `architecture_test.go` edge scan), strongly connected components, port order | todo | |
| 0.5 Go route dump export and `tools/route-dump` | todo | |
| 0.6 Spike: one `@http` raw route and one service trait implementation end to end | todo | |
| 0.7 media-engine Rust port with golden tests and benchmark | todo | |
| 0.8 `tools/parity` skeleton (throwaway MySQL, run Go and Rust, replay, diff) | todo | |

## Phase 1 Platform crates

| Unit | Status | Notes |
| --- | --- | --- |
| 1.1 `core` and config/env parity | todo | |
| 1.2 `db` | todo | |
| 1.3 `auth` (jwt, authorization, tenant, origins, securekey, emailcrypto) | todo | |
| 1.4 `events` (eventbus, jobs, emailqueue) | todo | |
| 1.5 `realtime` | todo | |
| 1.6 `integrations` | todo | |
| 1.7 `ai` (classifier) | todo | |
| 1.8 `server` (middleware chain, rate limit, Partials key bypass, health, version, error writer) | todo | |
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
| 6.2 Rust architecture check replaces `architecture_test.go`; update `AGENTS.md`, `CLAUDE.md`, docs | todo | |
| 6.3 Final verification pass, mark PR ready for review | todo | |

## Human action required

(Append infra, secrets, approvals, and pre-merge gate items here, with the commit that needs them.)
