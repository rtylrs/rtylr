/**
 * Cloudflare Worker — landing A/B variant router.
 *
 *   user -> rtylr.com -> THIS worker -> reads `rtylr_variant` cookie -> serves repo A or repo B
 *
 * The backend (`api.voxire.com/features`) decides the variant per visitor
 * (percentage rollout + targeting). The landing page writes it into the
 * `rtylr_variant` cookie on `.rtylr.com` (see apps/landing/lib/featureVariant.ts).
 * This worker only READS that cookie and routes — it never decides the variant.
 *
 * Bind two origins via Worker env vars / routes:
 *   REPO_A_ORIGIN — the current landing (this repo's deployment)   [default/control]
 *   REPO_B_ORIGIN — the new landing (repo-b), once it is built + deployed
 *
 * Until REPO_B_ORIGIN is set, everyone is served REPO_A_ORIGIN regardless of
 * the cookie, so this is safe to deploy now (it is a no-op pass-through until
 * repo-b exists). Route it on `rtylr.com/*`.
 */

const VARIANT_COOKIE = 'rtylr_variant';
const DEFAULT_VARIANT = 'a';

function readVariant(request) {
  const header = request.headers.get('Cookie') || '';
  for (const part of header.split(';')) {
    const [name, ...rest] = part.trim().split('=');
    if (name === VARIANT_COOKIE) {
      return decodeURIComponent(rest.join('=')) || DEFAULT_VARIANT;
    }
  }
  return DEFAULT_VARIANT;
}

export default {
  async fetch(request, env) {
    const variant = readVariant(request);

    const repoA = env.REPO_A_ORIGIN; // control / this repo (required)
    const repoB = env.REPO_B_ORIGIN; // repo-b (optional until it ships)

    // Choose the origin. Fail safe to A whenever B isn't configured/known.
    let origin = repoA;
    if (variant === 'b' && repoB) {
      origin = repoB;
    }

    // If no origins are configured yet, just pass through unchanged.
    if (!origin) {
      return fetch(request);
    }

    const incoming = new URL(request.url);
    const target = new URL(incoming.pathname + incoming.search, origin);

    const proxied = new Request(target.toString(), request);
    proxied.headers.set('X-Rtylr-Variant', variant);
    return fetch(proxied);
  },
};
