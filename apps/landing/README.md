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
