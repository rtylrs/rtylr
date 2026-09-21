# rtylr-landing

The RTYLR marketing site (rtylr.com), extracted from the `rtylr` monorepo into
its own repo so it builds, deploys, and evolves independently of the product
apps.

## Layout

```
apps/landing/      the Next.js (SSG) marketing site
packages/          git submodule -> voxire/packages (shared @voxire/* design system,
                   config, observability, consent, utils — consumed as workspace deps)
pnpm-workspace.yaml
```

## Setup

```bash
# clone WITH the packages submodule
git clone --recurse-submodules git@github.com:voxire/rtylr-landing.git
# (or, after a plain clone)
git submodule update --init --recursive

pnpm install
```

## Develop / build

```bash
pnpm dev        # next dev on :4200
pnpm typecheck  # tsc --noEmit
pnpm build      # next build (webpack) + next-sitemap -> dist/landing
```

## Deploy (Cloudflare Pages)

- Build command: `pnpm install && pnpm build`
- Output directory: `dist/landing`
- Enable git submodules in the Pages project (the build needs `packages/`).
- Node 20, pnpm 10.33.

## How a change reaches production

Nothing deploys on merge. Production ships only for a tag that
`tryrtylr/deploy` has been pointed at, and a tag exists only after a green run.

1. **Push to `main`.** Any change under `apps/landing/`, `packages/` or
   `pnpm-lock.yaml` starts the `Landing` workflow.
2. **Wait for `Landing` to go green.** The `checks` job runs typecheck, build,
   unit tests, the Open Graph image check and the end to end suite, in that
   order. `build` runs before the tests because both the unit suite and the
   image check read the static export out of `dist/landing`.
3. **A tag appears.** Once `checks` passes, the `version` job computes the next
   version from the existing `rtylr-landing-v*` tags and creates
   `rtylr-landing-v<version>`. A red run produces no tag, which is the gate:
   there is no branch protection here, so the tag is what stands between a
   broken build and production.
4. **Promote it.** From this checkout:

   ```bash
   scripts/promote-landing.sh rtylr-landing-v0.2.1
   ```

   That edits `apps/rtylr-landing.yaml` in `tryrtylr/deploy`, validates every
   manifest, and pushes. Add `--dry-run` first to see the diff without
   pushing. The script refuses a tag that is not on this repo's origin.
5. **Watch `deploy-release`** in this repo's Actions. `tryrtylr/deploy`
   dispatches it once the manifest change lands on its `main`.
6. **Verify.**

   ```bash
   curl -sI https://rtylr.com/ | grep -i cf-ray
   curl -s https://rtylr.com/version
   ```

   `cf-ray` confirms the response came from Cloudflare, and `/version` is
   served from `version.json` in the export, so it reflects the build that is
   actually live.

## Updating the shared packages

The `@voxire/*` packages live in the `packages` submodule. To pick up changes:

```bash
git submodule update --remote packages
git add packages && git commit -m "bump packages"
```

## A/B variant routing

`apps/landing/lib/featureVariant.ts` calls `api.voxire.com/features` on load and
writes the `rtylr_variant` cookie on `.rtylr.com`. The Cloudflare worker in
`apps/landing/cloudflare/variant-router.worker.js` reads that cookie and routes
to the A or B build (a no-op pass-through until `REPO_B_ORIGIN` is configured).
