/**
 * Runtime configuration client + store.
 *
 * Self-hosted RTYLR deployments ship the *prebuilt* frontend images, so the
 * apps cannot rely on the `*.rtylr.com` URLs baked in at build time — they must
 * learn their own URLs (and feature posture) from the backend at boot.
 *
 * The backend exposes an anonymous `GET {API_URL}/runtime-config` (also at
 * `/.well-known/rtylr.json`). We fetch it ONCE before the router/guards render,
 * cache it in a module-level singleton, and expose:
 *
 *   - `loadRuntimeConfig()` — the async boot fetch (idempotent, resilient).
 *   - `getRuntimeConfig()`  — a synchronous getter for the cached value.
 *   - resolvers (`resolveAppUrl`, `resolveAuthUrl`, `resolveApiBase`,
 *     `resolveCdnBaseUrl`) that return the *runtime* value when present and
 *     self-host is active, else the existing static default.
 *   - posture helpers (`isSelfHosted`, `allFeaturesEnabled`).
 *
 * HOSTED behavior is a strict no-op: in hosted deployments `self_hosted=false`
 * (and the URLs are the rtylr.com defaults), so every resolver falls through to
 * the pre-existing static value and nothing changes. A failed/blank fetch
 * likewise falls back to defaults — boot is never hard-blocked.
 */
import type { AppId, RuntimeConfig } from '@voxire/types';
import { APP_URLS } from './apps';

/**
 * Hosted CDN/upload default. Kept in sync with `DEFAULT_CDN_BASE_URL` in
 * `assets.ts` — duplicated here (rather than imported) so `assets.ts` can import
 * `resolveCdnBaseUrl` from this module without a circular dependency.
 */
const DEFAULT_CDN_BASE_URL = 'https://api.voxire.com';

interface RuntimeConfigState {
  /** Whether the boot fetch has resolved (success OR handled failure). */
  loaded: boolean;
  /** The fetched payload, or null when hosted/unavailable. */
  config: RuntimeConfig | null;
}

const state: RuntimeConfigState = { loaded: false, config: null };

let inflight: Promise<RuntimeConfig | null> | null = null;

/** Strip trailing slashes so we can safely concatenate paths. */
function stripTrailingSlash(url: string): string {
  return url.replace(/\/+$/, '');
}

/**
 * The API base used to *fetch* runtime config — the only URL we can know before
 * runtime config has loaded. Mirrors `@voxire/api`'s `getBaseUrl`: prefer the
 * build-time `VITE_API_URL`, else the hosted default. Note runtime config's own
 * `urls.api` then takes over for subsequent requests via `resolveApiBase`.
 */
function getBootstrapApiBase(): string {
  const env = (import.meta as ImportMeta & { env?: Record<string, string | undefined> }).env;
  return stripTrailingSlash(env?.VITE_API_URL || 'https://api.voxire.com/');
}

/**
 * True once the runtime config payload says this is a self-hosted deployment.
 * Hosted deployments report `self_hosted=false` (or omit it) → false.
 */
export function isSelfHosted(): boolean {
  const c = state.config;
  if (!c) return false;
  return c.self_hosted === true || c.deployment_mode === 'self_hosted';
}

/**
 * True when the backend says every app/feature should be treated as enabled
 * (self-host installs ship without the hosted licensing/upsell gating).
 */
export function allFeaturesEnabled(): boolean {
  return state.config?.all_features_enabled === true && isSelfHosted();
}

/** Synchronous accessor for the cached runtime config (null until loaded). */
export function getRuntimeConfig(): RuntimeConfig | null {
  return state.config;
}

/** Whether the boot fetch has completed (success or handled failure). */
export function isRuntimeConfigLoaded(): boolean {
  return state.loaded;
}

/**
 * Pick a runtime URL only when self-host is active and the value is a non-blank
 * string. Returns undefined otherwise so callers fall back to static defaults.
 */
function runtimeUrl(key: keyof NonNullable<RuntimeConfig['urls']>): string | undefined {
  if (!isSelfHosted()) return undefined;
  const value = state.config?.urls?.[key];
  if (typeof value !== 'string') return undefined;
  const trimmed = value.trim();
  return trimmed.length > 0 ? stripTrailingSlash(trimmed) : undefined;
}

/**
 * Resolve the absolute base URL for an app, preferring the runtime value in
 * self-host and falling back to the static `APP_URLS` default.
 */
export function resolveAppUrl(appId: AppId): string {
  const runtime = runtimeUrl(appId as keyof NonNullable<RuntimeConfig['urls']>);
  return runtime ?? APP_URLS[appId];
}

/** Resolve the auth app URL (used for SSO redirects, sign-out, 401). */
export function resolveAuthUrl(): string {
  return runtimeUrl('auth') ?? APP_URLS.auth;
}

/**
 * Resolve the API base. Falls back to the bootstrap base (VITE_API_URL / hosted
 * default) — which is what hosted always uses — when no runtime override exists.
 */
export function resolveApiBase(): string {
  return runtimeUrl('api') ?? getBootstrapApiBase();
}

/** Resolve the CDN/upload base used for asset reads + uploads. */
export function resolveCdnBaseUrl(): string {
  return runtimeUrl('upload') ?? DEFAULT_CDN_BASE_URL;
}

function hostnameOf(url: string | undefined): string | null {
  if (!url) return null;
  try {
    return new URL(url).hostname.toLowerCase();
  } catch {
    return null;
  }
}

/**
 * Cookie domain to use for the cross-app SSO token in SELF-HOST deployments.
 *
 * Hosted callers should NOT use this (it returns undefined unless self-host is
 * active) — they keep their existing `.rtylr.com` logic untouched.
 *
 * Self-host installs run on the customer's own domain(s). We derive the widest
 * shared parent domain across the configured app/auth/api hostnames so a single
 * token cookie is readable across every app subdomain. If the hosts share no
 * multi-label parent (e.g. a single-host install, or unrelated hosts), we return
 * undefined so the cookie stays host-only — which is the correct, safe default.
 */
export function resolveCookieDomain(): string | undefined {
  if (!isSelfHosted()) return undefined;
  const urls = state.config?.urls;
  if (!urls) return undefined;

  const appKeys: Array<keyof NonNullable<RuntimeConfig['urls']>> = [
    'auth',
    'dash',
    'pos',
    'erp',
    'hr',
    'crm',
    'finance',
    'flow',
    'insights',
    'recruit',
    'menu',
  ];

  const hosts = appKeys.map((k) => hostnameOf(urls[k])).filter((h): h is string => h !== null);

  if (hosts.length === 0) return undefined;

  // Longest common dotted suffix across all hosts.
  const split = hosts.map((h) => h.split('.').reverse());
  const first = split[0];
  const common: string[] = [];
  for (let i = 0; i < first.length; i++) {
    const label = first[i];
    if (split.every((parts) => parts[i] === label)) {
      common.push(label);
    } else {
      break;
    }
  }
  // Need at least a two-label registrable parent (e.g. example.com) to scope a
  // shared cookie. A single label (e.g. just "localhost") is not shareable.
  if (common.length < 2) return undefined;
  const domain = common.slice().reverse().join('.');
  return `.${domain}`;
}

/**
 * Fetch and cache `/runtime-config` exactly once. Safe to call repeatedly: the
 * first call performs the request, later calls return the same promise (or the
 * already-cached value). Never throws — a failure leaves the cache empty so all
 * resolvers fall back to static defaults.
 */
export function loadRuntimeConfig(): Promise<RuntimeConfig | null> {
  if (state.loaded) return Promise.resolve(state.config);
  if (inflight) return inflight;

  inflight = (async (): Promise<RuntimeConfig | null> => {
    try {
      const base = getBootstrapApiBase();
      const res = await fetch(`${base}/runtime-config`, {
        method: 'GET',
        headers: { Accept: 'application/json' },
        credentials: 'omit',
      });
      if (res.ok) {
        const data = (await res.json()) as RuntimeConfig;
        if (data && typeof data === 'object') {
          state.config = data;
        }
      }
    } catch {
      // Network/parse failure — stay on static defaults. Never hard-block boot.
    } finally {
      state.loaded = true;
      inflight = null;
    }
    return state.config;
  })();

  return inflight;
}
