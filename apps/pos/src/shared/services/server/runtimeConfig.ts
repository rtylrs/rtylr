import { normalizeBaseUrl } from './serverConfig';

/**
 * A self-hosted rtylr backend may expose `GET /runtime-config` returning the
 * public URLs its frontends should use:
 *   { urls: { auth, upload, ... } }
 * We fetch this once when the user configures a custom server so the app can
 * redirect to the customer's own auth host and upload endpoint.
 */
export interface RuntimeConfigUrls {
  auth?: string;
  upload?: string;
  [key: string]: string | undefined;
}

export interface RuntimeConfigResult {
  authUrl: string | null;
  uploadUrl: string | null;
}

/**
 * Fetch `/runtime-config` from the given API base. Never throws — on any
 * failure it resolves to nulls so the caller can fall back to derived URLs.
 * Bounded by a 5s timeout: a hung self-host server must not stall startup.
 */
export async function fetchRuntimeConfig(apiBaseUrl: string): Promise<RuntimeConfigResult> {
  const base = normalizeBaseUrl(apiBaseUrl);
  try {
    const res = await fetch(`${base}runtime-config`, {
      method: 'GET',
      headers: { Accept: 'application/json' },
      signal: AbortSignal.timeout(5_000),
    });
    if (!res.ok) return { authUrl: null, uploadUrl: null };
    const body = (await res.json()) as { urls?: RuntimeConfigUrls };
    const urls = body.urls ?? {};
    return {
      authUrl: typeof urls.auth === 'string' ? urls.auth : null,
      uploadUrl: typeof urls.upload === 'string' ? urls.upload : null,
    };
  } catch (error) {
    console.warn('[runtime-config] fetch failed; using derived URLs', error);
    return { authUrl: null, uploadUrl: null };
  }
}
