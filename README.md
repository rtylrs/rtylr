# rtylr

The rtylr product in one repository. One checkout, one install, one place to
change a contract and every consumer with it.

## Layout

| Path | What it is |
|---|---|
| `apps/rtylr/` | The rtylr workspace: the ten product apps (`dash`, `auth`, `crm`, `erp`, `finance`, `flow`, `hr`, `insights`, `menu`, `recruit`) plus their tests, docs and scripts. |
| `apps/landing/` | The marketing site (Next.js static export). |
| `apps/pos/` | The POS terminal: Vite web app plus the Tauri desktop shell in `core/`. |
| `apps/docs/` | The documentation site (Next.js static export). |
| `apps/mcp/` | The MCP server. Build-only; nothing deploys it. |
| `packages/` | The shared library layer (`@voxire/*`). |
| `voxie/` | API contracts (`.onk`) plus generated Go, TypeScript and OpenAPI output. In-tree and the single source of truth. |
| `services/onion/` | The Go backend: modules, platform packages, sqlc, worker. |

`tryrtylr/deploy`, `tryrtylr/infra` and `tryrtylr/pipelines` stay separate:
they are the release control plane, the infrastructure, and the reusable
workflows.

## One workspace

Everything under `apps/` and `packages/` is a member of the pnpm workspace at
the repository root, with a single lockfile. Apps depend on the shared layer
through `workspace:*`.

This is deliberate. Separate installs for the same packages gave a linked
package one copy of its dependencies per consumer, and type identity needs the
same declaration files, not the same versions.

```sh
pnpm install          # installs every workspace member
pnpm -r typecheck     # every member
pnpm --filter @rtylr/dash test
pnpm --filter dash dev
pnpm -C apps/rtylr build   # the ten rtylr apps, one command
```

The rtylr workspace keeps its own scripts (`apps/rtylr/package.json`) for the
things that only concern it: the contracts guard, contracts, the vitest
suite across packages and apps, and the app build fan-out.

## Releases and deployment

Nothing in this repository deploys itself:

1. A push to `main` cuts release metadata only — one Version workflow per
   product (`.github/workflows/ci.yml`, `landing.yml`, `pos.yml`, `docs.yml`),
   each path-filtered, each with its own tag namespace (`rtylr-v*`, `pos-v*`,
   `rtylr-landing-v*`, `rtylr-docs-v*`).
2. A maintainer promotes one exact release in `tryrtylr/deploy` by editing the
   app's manifest.
3. `deploy` dispatches back here; `.github/workflows/deploy-release.yml` builds
   the promoted tag through `tryrtylr/pipelines` and ships it.

Manifests carry `working_directory` (for example `apps/rtylr`) plus `dist_dir`,
so one adapter serves every product.

Other workflows:

- `desktop-release.yml` — Tauri desktop builds and the public downloads site.
- `publish-images.yml` / `publish-images-pos.yml` — the public self-host images,
  triggered by a version bump in `tryrtylr/rtylr-version`.

## Development notes

- `apps/landing` and `apps/docs` export statically; their build output stays
  inside the app (`dist/landing`, `out`).
- `apps/pos` builds the web app into `apps/pos/dist`; the desktop shell builds
  through Tauri from `apps/pos/core`.
- Contract output is generated, not committed. Run `make -C voxie generate`
  after changing anything under `voxie/api/`; CI regenerates before every build.
  `voxie/` is the Go module `github.com/voxire/voxie`, consumed by `onion`
  through a `replace` directive and by the apps as the `@voxire/voxie`
  workspace package.
