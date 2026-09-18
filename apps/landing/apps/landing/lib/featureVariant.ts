import api from './api';
import { getCookie, setCookie, getDefaultCookieDomain } from './cookies';

// Cookie the Cloudflare worker reads to decide which build/repo to serve.
const VARIANT_COOKIE = 'rtylr_variant';
// Stable anonymous id so the percentage rollout is sticky across visits.
const ANON_COOKIE = 'rtylr_aid';

export const DEFAULT_VARIANT = 'a';

const ANON_ID_DAYS = 365;
const VARIANT_DAYS = 30;

export interface FeatureResolution {
  key: string;
  enabled: boolean;
  variant: string;
}

export interface FeaturesResponse {
  variant?: string;
  features?: Record<string, FeatureResolution>;
}

let landingFeaturesPromise: Promise<FeaturesResponse> | null = null;

function getOrCreateAnonId(): string {
  const existing = getCookie(ANON_COOKIE);
  if (existing) return existing;
  const id =
    typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID()
      : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
  setCookie(ANON_COOKIE, id, { days: ANON_ID_DAYS, domain: getDefaultCookieDomain() });
  return id;
}

function writeVariant(variant: string): void {
  setCookie(VARIANT_COOKIE, variant, { days: VARIANT_DAYS, domain: getDefaultCookieDomain() });
}

function fallbackFeatures(): FeaturesResponse {
  return { variant: DEFAULT_VARIANT, features: {} };
}

export async function resolveLandingFeatures(): Promise<FeaturesResponse> {
  if (!landingFeaturesPromise) {
    landingFeaturesPromise = (async () => {
      try {
        const subject = getOrCreateAnonId();
        const { data } = await api.post<FeaturesResponse>('/features', { subject });
        return {
          variant: data?.variant || DEFAULT_VARIANT,
          features: data?.features || {},
        };
      } catch {
        return fallbackFeatures();
      }
    })();
  }

  return landingFeaturesPromise;
}

/**
 * Resolves the landing variant from the backend feature-support service and
 * writes it to the `rtylr_variant` cookie on `.rtylr.com`, so the Cloudflare
 * worker can route the next request to the matching build. The flow:
 *
 *   user -> rtylr.com -> worker (reads rtylr_variant) -> serves variant a|b
 *
 * On the very first visit there is no cookie yet, so the worker serves the
 * default (variant a = this repo) and this call seeds the cookie for next time.
 * Any failure (backend down, network, CORS) falls back to variant 'a' so the
 * landing never breaks.
 */
export async function initFeatureVariant(): Promise<string> {
  const features = await resolveLandingFeatures();
  const variant = features.variant || DEFAULT_VARIANT;
  writeVariant(variant);
  return variant;
}
