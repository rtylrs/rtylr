/**
 * Backend selection resolver.
 *
 * The POS app normally talks to the hosted rtylr backend, but a customer
 * running a self-hosted deployment can point the app at their own server.
 * The chosen backend is persisted by `serverStore` (Zustand + persist) under
 * the localStorage key below.
 *
 * These resolvers are deliberately framework-free and read localStorage
 * synchronously, because the axios client (`api.ts`) and the realtime socket
 * (`realtime.ts`) resolve their base URLs at module-init time — before React
 * (and therefore the Zustand store) has hydrated.
 *
 * Resolution order for the API base URL:
 *   1. self-host URL from the persisted store (mode === 'self_host')
 *   2. import.meta.env.VITE_API_URL
 *   3. the hosted default
 */

export type ServerMode = 'hosted' | 'self_host';

export const SERVER_STORE_KEY = 'pos-server-config';

export const HOSTED_API_URL = 'https://api.voxire.com/';
export const HOSTED_AUTH_URL = 'https://sso.rtylr.com';

export function isHostedPosOrigin(): boolean {
  if (typeof window === 'undefined') return false;
  return window.location.hostname.toLowerCase() === 'pos.rtylr.com';
}

interface PersistedServerState {
  mode: ServerMode;
  customApiUrl: string | null;
  authUrl: string | null;
  uploadUrl: string | null;
}

function readPersistedState(): PersistedServerState | null {
  if (isHostedPosOrigin()) {
    return {
      mode: 'hosted',
      customApiUrl: null,
      authUrl: null,
      uploadUrl: null,
    };
  }

  if (typeof localStorage === 'undefined') return null;
  try {
    const raw = localStorage.getItem(SERVER_STORE_KEY);
    if (!raw) return null;
    // zustand/persist wraps state as { state, version }.
    const parsed = JSON.parse(raw) as { state?: Partial<PersistedServerState> };
    const state = parsed.state;
    if (!state || typeof state !== 'object') return null;
    return {
      mode: state.mode === 'self_host' ? 'self_host' : 'hosted',
      customApiUrl: typeof state.customApiUrl === 'string' ? state.customApiUrl : null,
      authUrl: typeof state.authUrl === 'string' ? state.authUrl : null,
      uploadUrl: typeof state.uploadUrl === 'string' ? state.uploadUrl : null,
    };
  } catch {
    return null;
  }
}

/** Ensure a base URL ends with a single trailing slash. */
export function normalizeBaseUrl(url: string): string {
  return `${url.trim().replace(/\/+$/, '')}/`;
}

/** Validate a user-entered server URL. Must be an absolute http(s) URL. */
export function isValidServerUrl(value: string): boolean {
  try {
    const u = new URL(value.trim());
    return u.protocol === 'http:' || u.protocol === 'https:';
  } catch {
    return false;
  }
}

const envApiUrl = (): string | undefined =>
  (import.meta as ImportMeta & { env: Record<string, string> }).env.VITE_API_URL || undefined;

/**
 * Resolve the API base URL: self-host store → VITE_API_URL → hosted default.
 * Always returns a value ending in a trailing slash.
 */
export function resolveApiBaseUrl(): string {
  const state = readPersistedState();
  if (state?.mode === 'self_host' && state.customApiUrl) {
    return normalizeBaseUrl(state.customApiUrl);
  }
  return normalizeBaseUrl(envApiUrl() ?? HOSTED_API_URL);
}

/** Resolve the websocket base URL derived from the API base. */
export function resolveWsBaseUrl(): string {
  const base = resolveApiBaseUrl().replace(/\/$/, '');
  const wsScheme = base.startsWith('https://') ? 'wss://' : 'ws://';
  return `${wsScheme}${base.replace(/^https?:\/\//, '')}`;
}

/** Derive a fallback auth URL from a self-hosted API URL when runtime-config is unavailable. */
export function deriveSelfHostAuthUrl(apiUrl: string): string {
  try {
    const url = new URL(apiUrl);
    const hostname = url.hostname;

    // 1. Subdomain replacements: e.g. api.domain.com -> auth.domain.com
    if (hostname.startsWith('api.')) {
      url.hostname = 'auth.' + hostname.slice(4);
      return url.origin;
    }
    if (hostname.includes('api')) {
      const newHostname = hostname
        .replace(/\bapi\b/g, 'auth')
        .replace(/-api/g, '-auth')
        .replace(/api-/g, 'auth-');
      if (newHostname !== hostname) {
        url.hostname = newHostname;
        return url.origin;
      }
    }

    // 2. Path replacement: e.g. domain.com/api/v1 -> domain.com/auth
    const pathname = url.pathname.replace(/\/+$/, '');
    const apiIndex = pathname.indexOf('/api');
    if (apiIndex !== -1) {
      return `${url.origin}${pathname.substring(0, apiIndex)}/auth`;
    }

    // 3. General fallback: append /auth
    return `${url.origin}/auth`;
  } catch {
    return HOSTED_AUTH_URL;
  }
}

/**
 * Resolve the auth (sign-in) base URL. For self-host this comes from the
 * customer backend's runtime-config when known; otherwise we fall back to the
 * hosted auth host so the app stays usable.
 */
export function resolveAuthUrl(): string {
  const state = readPersistedState();
  if (state?.mode === 'self_host') {
    if (state.authUrl) return state.authUrl.replace(/\/$/, '');
    if (state.customApiUrl) return deriveSelfHostAuthUrl(state.customApiUrl);
    return HOSTED_AUTH_URL;
  }
  return HOSTED_AUTH_URL;
}
