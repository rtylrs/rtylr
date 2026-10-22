# rtylr

The rtylr product in one repository. One checkout, one install, one place to
change a contract and every consumer with it.

## Layout

| Path                   | What it is                                                                                                                    |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `clients/`             | The pnpm workspace root. Everything JavaScript/TypeScript lives under it.                                                     |
| `clients/app/`         | Single Portal application: business modules, the public customer menu, and separate SSO and public Recruit build entrypoints. |
| `clients/web/landing/` | The marketing site (Next.js static export).                                                                                   |
| `clients/web/docs/`    | The documentation site (Next.js static export).                                                                               |
| `clients/mcp/`         | The MCP server. Build-only; nothing deploys it.                                                                               |
| `clients/pos/`         | The POS terminal: Vite web app plus the Tauri desktop shell in `core/`.                                                       |
| `clients/packages/`    | The shared library layer (`@voxire/*`) and the generated contracts package (`contracts/`).                                    |
| `voxie/`               | API contracts (`.onk`) plus the generated Go and OpenAPI output. In-tree and the single source of truth.                      |
| `services/onion/`      | The Go backend: modules, platform packages, sqlc, worker.                                                                     |

`rtylrs/infra` and `rtylrs/pipelines` stay separate: they are the
infrastructure and the reusable build-and-deploy workflows.

## One workspace

Everything under `clients/` is a member of the pnpm workspace rooted at
`clients/`, with a single lockfile. Apps depend on the shared layer through
`workspace:*`.

This is deliberate. Separate installs for the same packages gave a linked
package one copy of its dependencies per consumer, and type identity needs the
same declaration files, not the same versions.

```sh
pnpm -C clients install          # installs every workspace member
pnpm -C clients -r typecheck     # every member
pnpm -C clients --filter @rtylr/portal test
pnpm -C clients --filter @rtylr/portal dev
pnpm -C clients/app build  # Portal, SSO and careers builds
```

The rtylr workspace keeps its own scripts (`clients/app/package.json`)
for the things that only concern it: the contracts guard, contracts, the vitest
suite across packages and apps, and the app build fan-out.

## Setup

```sh
pnpm -C clients install --frozen-lockfile
make -C voxie generate   # contracts are generated, not committed
git config core.hooksPath .githooks
```

The last line wires the committed hooks; git does not track `core.hooksPath`.

## Releases and deployment

A push to `main` releases and deploys:

1. `.github/workflows/release.yml` computes the next `rtylr-v*` version,
   attaches only the components that changed (per-component source archives
   plus a `release.json` manifest), and creates the tag and GitHub release.
2. The same run picks the deploy targets whose own paths changed, using
   `.github/deploy-targets.json`, and `.github/workflows/deploy.yml` builds the
   new tag and ships it — Pages apps and ECS services through the same
   entrypoint. ECS images build while the database migrates, and the services
   roll out once both are done. A change that touches no target's paths (CI
   only, for example) cuts a release and deploys nothing.
3. To re-ship or roll back an existing tag, run `redeploy.yml` with the tag and
   a target name (or `all`).

Each target in `deploy-targets.json` lists the path prefixes that trigger it
(with optional `ignore` prefixes) and carries `working_directory` (for example
`clients/app`) plus `dist_dir`, so one entrypoint serves every product. Adding an app is one entry in that file.

A release or redeploy that ships the API or the worker first applies its
database migrations with `migrate-onion-database.yml`, unless that exact set
(`.github/scripts/migration-key.sh`) already applied cleanly. Run that workflow
by hand to rehearse a migration (`dry_run`) or to apply one again.

Other workflows:

- `desktop-release.yml` — Tauri desktop builds and the public downloads site;
  installer releases are still tagged `pos-v*` and are separate from source
  releases.

## Development notes

- `clients/web/landing` and `clients/web/docs` export statically; their build
  output stays inside the app (`dist/landing`, `out`).
- `clients/pos` builds the web app into `clients/pos/dist`; the desktop shell
  builds through Tauri from `clients/pos/core`.
- Contract output is generated, not committed. Run `make -C voxie generate`
  after changing anything under `voxie/api/`; CI regenerates before every build.
  `voxie/` is the Go module `github.com/voxire/voxie`, consumed by `onion`
  through a `replace` directive; its TypeScript half is copied into
  `clients/packages/contracts`, which the apps consume as `@voxire/voxie`.

## License

Everything in this repository is source-available under the
[PolyForm Strict License 1.0.0](./LICENSE). It permits noncommercial use only
and does not grant the right to distribute the software or make changes or new
works based on it. Commercial use, self-hosting, redistribution and
modification require a separate written license from Voxire.

This single LICENSE at the repository root covers every path. PolyForm Strict
has no SPDX license identifier, so do not invent one in package manifests, and
any artifact distributed outside this repository — an npm tarball, a Go module
zip, a container image — must include a copy of this LICENSE.
