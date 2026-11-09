# Agent prompt: continue the Onion Go to Rust migration on this PR

Paste everything below the line into opencode. It is resumable: run it again after any stop and it continues from `STATE.md`.

---

You are continuing a migration of the Go backend in `services/onion` (monorepo `rtylrs/rtylr`) to Rust. The whole migration lives in ONE long-lived branch, `rust-migration`, and ONE draft pull request. The plan is already designed and decided. Do not redesign it, do not split it into more PRs, and do not propose a gradual or strangler rollout.

## Where you work

Check out branch `rust-migration` (a worktree under `.claude/worktrees/` or a clone). Run `git pull --ff-only` first. Never work on `main`. All paths below are relative to the repo root.

## Read first, in this order, before touching anything

Plan and state:
1. `services/onion/docs/rust-migration/ARCHITECTURE.md` (source of truth)
2. `services/onion/docs/rust-migration/STATE.md` (task list and progress)
3. `AGENTS.md` and `services/onion/CLAUDE.md` (repo rules, ownership conventions, env vars, migration safety)

Backend entrypoints and wiring (read when a unit touches them):
4. `services/onion/cmd/api/main.go`, `main_modules.go`, `main_routes.go`, `module_registrars.go`, `compat_routes.go`, `route_dump_test.go`, `compat_parity_test.go`
5. `services/onion/cmd/worker/main.go`
6. `services/onion/architecture_test.go` (import-edge scan, reuse for the coupling graph)
7. `services/onion/internal/platform/server/http/web_server.go` (middleware chain, rate limit, Partials key bypass, origins)
8. `services/onion/internal/platform/config/config.go`, `config_part2.go`, `services/onion/.env.example`
9. `services/onion/internal/platform/jobs/` (`registry.go`, `scheduler.go`, `mysql_locker.go`) and `internal/platform/eventbus/`
10. `services/onion/internal/platform/cdn/engine.go` and `services/onion/media-engine/` (CMakeLists, `src`, `include`)
11. `services/onion/internal/ai/classifier/` (checkpoint, model, planner, tests, testdata)

Contracts and generation:
12. `voxie/onekit.toml`, `voxie/Makefile`, and a sample contract such as `voxie/api/hr/time_entry/v1/`. `onek` is pinned at v0.36.3 and supports `rust-server`. Never edit generated output.

Scripts, build, and delivery (read when a unit touches them):
13. `services/onion/Makefile`, `scripts/migrate.sh`, `scripts/test-db.sh`, `scripts/generate_sqlc.sh`, `scripts/build-binary.sh`, `scripts/check_env_example.py`, `scripts/check-route-contract-drift.py`, `scripts/check_schema_drift.py`, `scripts/rust-capped.sh`
14. `services/onion/Dockerfile`, `Dockerfile.worker`, `Dockerfile.migrate`
15. `.github/deploy-targets.json` and `.github/workflows/`: `release.yml`, `deploy.yml`, `build-ecs-image.yml`, `deploy-ecs-release.yml`, `go-build-cache.yml`, `onion-images.yml`, `migrate-onion-database.yml`, `redeploy.yml`, `opencode-review.yml`; plus `.github/scripts/build-onion-binaries.sh` and `.github/scripts/migration-key.sh`
16. Other backend docs when relevant: `services/onion/docs/backend.md`, `platform-patterns.md`, `database.md`, `sqlc-query-ownership.md`, `tenant-index-matrix.md`

## Your loop

Repeat until `STATE.md` has no `todo`, `in-progress`, or unblocked rows, or a stop condition below fires:

1. Pick the first row in `STATE.md` that is `in-progress`, else the first `todo` whose prerequisites are `done`. Phase order is binding; within Phase 3 follow the order in `COUPLING.md` once it exists.
2. Do exactly that unit: port behavior 1:1, write or port tests first or alongside, run the verification below.
3. Update `STATE.md` (status, notes) and commit everything for that unit in ONE commit: `feat(rust): <unit id> <what>`. Commit with the identity already configured; do not touch git config.
4. Push only at the end of each phase, or to fix failing CI. Do not push after every commit: the review bot has a small daily quota. Then run `gh pr checks` to see CI.
5. Next unit.

If this is the very first run and no PR exists, open it as a DRAFT with `gh pr create --draft` after unit 0.1 is pushed, with a short description pointing at `ARCHITECTURE.md`. Otherwise use the existing PR. Keep the PR a draft until Phase 6 is `done`; only then run `gh pr ready`.

## Hard rules (violating any means stop and report)

- Never connect to a production or shared database. The main checkout's `.env` points at production RDS. Use only the throwaway local MySQL from `scripts/test-db.sh` (container `onion-env-db-1`). If you cannot get one, mark the unit `blocked` and move to a unit that needs none.
- Never use the local `aws` CLI. Never touch AWS, `rtylrs/infra`, `rtylrs/pipelines`, or trigger any workflow (`gh workflow run`, `redeploy`, deploy workflows). Editing the workflow and deploy-target files in this repo is allowed only in Phase 5 and 6 and only as the units say; they take effect on merge. Anything needing infra goes under "Human action required" in `STATE.md`.
- Never merge the PR, deploy, tag, or publish. Never mark the PR ready before Phase 6 is done.
- Never change git identity or config. Never force-push. Never rewrite pushed history.
- Never copy secrets, tokens, or keys into code, tests, fixtures, logs, or docs. Use synthetic data only. Do not read `.env` files other than `.env.example`.
- Never edit or commit generated output (`voxie/gen/`, `crates/contracts/src/generated/`, sqlc output). Contract changes only in `voxie/api`, and only if a unit requires it.
- Do not upgrade `onek` (stay on v0.36.3). Do not add paid CI or paid runners; free GitHub-hosted runners and Actions cache only.
- Do not change Go behavior except where a unit says so. Do not refactor Go. Go stays working until unit 6.1 deletes it.
- Do not move SQL files. `internal/modules/*/sql` and `internal/platform/*/sql` stay exactly where they are; the migrate image and migration workflows read them.
- Do not invent behavior. If Go is ambiguous, port what it does and note it in `STATE.md`. Environment variable names and semantics stay identical to Go. Port ownership columns exactly as Go does (e.g. `hr_time_entries.owner_id` is the business ID).
- Do not use `unwrap`/`expect`/`panic` outside tests and startup. No `f64` for money. No global singletons. Comments only where needed, a few lines at most. Do not add dependencies outside the table in `ARCHITECTURE.md` section 5 without a one-line reason in the commit message.
- Do not spawn subagents or run anything in parallel. One thread, one unit at a time.

## Resource caps (this server is small; these are hard rules)

- Run EVERY heavy command through `services/onion/scripts/rust-capped.sh`: cargo, go, make, docker, `scripts/test-db.sh`. Example: `services/onion/scripts/rust-capped.sh cargo check -p core`. It enforces a hard memory cap and CPU quota via a systemd cgroup and fails closed if it cannot. If it refuses to run, stop and report; do not set `RM_ALLOW_UNCAPPED` and do not raise `RM_MEM_MAX`/`RM_CPUS`.
- Never run `docker build` or `cargo build --release` locally. CI builds images (`onion-images.yml`).
- Prefer `cargo check -p <crate>` and `cargo test -p <crate>`; run the full workspace check only at the end of a phase. Go tests only for the package you are comparing against (`go test ./internal/modules/<m>/...`), never `./...`.
- Cap the throwaway MySQL before first use: `docker update --memory 1g --memory-swap 1g --cpus 1 onion-env-db-1`.
- If a command dies (exit 137, OOM kill, timeout, `No space left`), shrink the scope (`-p`, `-j 1`, fewer tests, `cargo clean -p <crate>`) and retry. Never raise caps. If it still fails, mark the unit `blocked` with the exact command and error.
- Free disk below 5 GB means stop and report.

## Verification you must run and report for every unit

Through the capped wrapper, for the crates the unit touched:

```
cargo fmt --check
cargo clippy -p <crate> --all-targets -- -D warnings
cargo check -p <crate>
cargo test -p <crate>          # DB tests only against the throwaway MySQL
```

At the end of every phase also run `cargo check --workspace` and `cargo clippy --workspace -- -D warnings`.

For each Phase 3 module also run: route parity (`tools/route-dump` vs the Go route table), the HTTP replay diff (`tools/parity`), the module's ported DB tests, and `python3 services/onion/scripts/check_schema_drift.py`, `check_env_example.py`, `check-route-contract-drift.py`. "Zero tests collected" or "no projects matched" is not a pass. If something cannot run, say so and why; never claim it passed. If a Go test fails, confirm it fails on `main` first and record it as a baseline failure.

## Stop and report (do not guess) when

- a unit needs a production credential, a shared database, or AWS;
- Go behavior is unclear in a way that affects money, authorization, tenant isolation, or lock keys and job names;
- a dependency has no workable crate and no fallback in the plan;
- tests show Go behavior contradicts `ARCHITECTURE.md`;
- a prerequisite row is not actually done;
- the capped wrapper or disk guard stops you and shrinking scope does not help.

When you stop, set the row to `blocked` with the exact reason, append a line to "Human action required" if a human must act, commit the STATE update, and report.

## Final message format (every stop, including normal completion)

Report: units completed this run (ids and commits), verification commands with results, what you could not verify, new "Human action required" items, current phase and next unit, and CI status of the PR. When Phase 6 is done: confirm the pre-merge gate in `ARCHITECTURE.md` section 11 is the human's next step, and do not merge.
