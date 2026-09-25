# rtylr

The rtylr product in one repository. One checkout, one install, one place to
change a contract and every consumer with it.

## Layout

| Path | What it is |
|---|---|
| `clients/` | The pnpm workspace root. Everything JavaScript/TypeScript lives under it. |
| `clients/web/rtylr/` | The rtylr suite: the ten product apps (`dash`, `auth`, `crm`, `erp`, `finance`, `flow`, `hr`, `insights`, `menu`, `recruit`) plus their tests, docs and scripts. |
| `clients/web/landing/` | The marketing site (Next.js static export). |
| `clients/web/docs/` | The documentation site (Next.js static export). |
| `clients/web/mcp/` | The MCP server. Build-only; nothing deploys it. |
| `clients/pos/` | The POS terminal: Vite web app plus the Tauri desktop shell in `core/`. |
| `clients/packages/` | The shared library layer (`@voxire/*`) and the generated contracts package (`contracts/`). |
| `voxie/` | API contracts (`.onk`) plus the generated Go and OpenAPI output. In-tree and the single source of truth. |
| `services/onion/` | The Go backend: modules, platform packages, sqlc, worker. |

`tryrtylr/deploy`, `tryrtylr/infra` and `tryrtylr/pipelines` stay separate:
they are the release control plane, the infrastructure, and the reusable
workflows.

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
pnpm -C clients --filter @rtylr/dash test
pnpm -C clients --filter dash dev
pnpm -C clients/web/rtylr build  # the ten rtylr apps, one command
```

The rtylr workspace keeps its own scripts (`clients/web/rtylr/package.json`)
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

Nothing in this repository deploys itself:

1. A push to `main` cuts one release — `.github/workflows/release.yml` computes
   the next `rtylr-v*` version, attaches only the components that changed
   (per-component source archives plus a `release.json` manifest), and creates
   the tag and GitHub release.
2. A maintainer promotes one exact release in `tryrtylr/deploy` by editing the
   app's manifest.
3. `deploy` dispatches back here; `.github/workflows/deploy-release.yml` builds
   the promoted tag through `tryrtylr/pipelines` and ships it — Pages apps and
   ECS services through the same adapter.

Manifests carry `working_directory` (for example `clients/web/rtylr`) plus
`dist_dir`, so one adapter serves every product.

Other workflows:

- `desktop-release.yml` — Tauri desktop builds and the public downloads site;
  installer releases are still tagged `pos-v*` and are separate from source
  releases.
- `publish-images.yml` / `publish-images-pos.yml` — the public self-host images,
  dispatched by the Release workflow when app sources change (or run manually).

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
