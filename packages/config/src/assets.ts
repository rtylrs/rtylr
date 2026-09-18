import { resolveCdnBaseUrl } from './runtimeConfig';

export const ASSETS = {
  // Served from each app's own origin (public/logo.png) rather than an external
  // CDN — the previous cdn.voxire.dev host no longer resolves. The asset is a
  // single near-white mark; consumers tint it per background (see the auth
  // Logo component).
  logos: {
    dark: '/logo.png',
    light: '/logo.png',
  },
  /** Per-app logos - will be populated when individual logos are designed.
   *  Falls back to the main rtylr logo for now. */
  appLogos: {
    hub: { dark: '', light: '' },
    pos: { dark: '', light: '' },
    erp: { dark: '', light: '' },
    hr: { dark: '', light: '' },
    crm: { dark: '', light: '' },
    finance: { dark: '', light: '' },
    flow: { dark: '', light: '' },
    insights: { dark: '', light: '' },
    menu: { dark: '', light: '' },
    auth: { dark: '', light: '' },
  } as Record<string, { dark: string; light: string }>,
};

export interface CdnImageUrlOptions {
  cdnBaseUrl?: string;
}

export const DEFAULT_CDN_BASE_URL = 'https://api.voxire.com';

const CDN_READ_PATHS = ['/files/', '/images/'];
const LOCAL_IMAGE_SCHEMES = /^(?:asset|blob|data):/i;
const ABSOLUTE_URL = /^[a-z][a-z0-9+.-]*:/i;
const CDN_FILE_ID = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
const VOXIRE_CDN_HOSTS = new Set(['cdn.voxire.com', 'cdn.voxire.dev', 'api.voxire.com', 'api.voxire.dev']);

export function getCdnBaseUrl(override?: string): string {
  // In self-host, `resolveCdnBaseUrl()` returns the runtime upload base; hosted
  // (and a blank runtime config) returns the same `DEFAULT_CDN_BASE_URL`, so the
  // hosted resolution is byte-for-byte identical to before.
  return (override || resolveCdnBaseUrl() || DEFAULT_CDN_BASE_URL).replace(/\/+$/, '');
}

/** Recognised CDN hosts: the static voxire hosts plus the runtime upload host. */
function recognisedCdnHosts(): Set<string> {
  const hosts = new Set(VOXIRE_CDN_HOSTS);
  try {
    const runtime = resolveCdnBaseUrl();
    if (runtime) hosts.add(new URL(runtime).hostname.toLowerCase());
  } catch {
    // ignore malformed runtime base
  }
  return hosts;
}

export function cdnImageUrl(
  value: string | null | undefined,
  options: CdnImageUrlOptions = {}
): string | undefined {
  const raw = typeof value === 'string' ? value.trim() : '';
  if (!raw) return undefined;
  if (LOCAL_IMAGE_SCHEMES.test(raw)) return raw;

  const base = getCdnBaseUrl(options.cdnBaseUrl);
  if (CDN_FILE_ID.test(raw)) {
    return `${base}/upload?id=${encodeURIComponent(raw)}`;
  }
  if (raw.startsWith('files/') || raw.startsWith('images/')) {
    return `${base}/${raw}`;
  }
  const canonicalPath = CDN_READ_PATHS.some((prefix) => raw.startsWith(prefix));
  if (canonicalPath || raw === '/upload' || raw.startsWith('/upload?')) {
    return `${base}${raw}`;
  }

  if (ABSOLUTE_URL.test(raw)) {
    try {
      const parsed = new URL(raw);
      const host = parsed.hostname.toLowerCase();
      const isVoxireCDN = recognisedCdnHosts().has(host);
      const isCanonicalCDNPath =
        CDN_READ_PATHS.some((prefix) => parsed.pathname.startsWith(prefix)) ||
        parsed.pathname === '/upload';
      if (isVoxireCDN && isCanonicalCDNPath) {
        return `${base}${parsed.pathname}${parsed.search}`;
      }
    } catch {
      // Fall through and let the upload runmode reject malformed URLs.
    }
  }

  return `${base}/upload?url=${encodeURIComponent(raw)}`;
}

/** Returns the logo URL for the current theme. Falls back to main rtylr logo. */
export function getLogoUrl(appId?: string, theme: 'dark' | 'light' = 'light'): string {
  if (appId && ASSETS.appLogos[appId]) {
    const appLogo = ASSETS.appLogos[appId][theme];
    if (appLogo) return appLogo;
  }
  return ASSETS.logos[theme];
}

export const AVATAR_STYLES = ['thumbs', 'bottts-neutral', 'fun-emoji'] as const;
export type AvatarStyle = (typeof AVATAR_STYLES)[number];

function hashSeed(s: string): number {
  let h = 0;
  for (let i = 0; i < s.length; i++) h = (h * 31 + s.charCodeAt(i)) >>> 0;
  return h;
}

export function pickAvatarStyle(seed?: string | null): AvatarStyle {
  const key = (seed || 'user').trim().toLowerCase();
  return AVATAR_STYLES[hashSeed(key) % AVATAR_STYLES.length];
}

export function getAvatarUrl(seed?: string | null, style?: AvatarStyle): string {
  const raw = (seed || 'user').trim().toLowerCase();
  // The seed is usually a user's email/name. Hashing it keeps the raw PII
  // *string* out of the third-party avatar URL (so an email isn't sitting in
  // dicebear's request logs / the Referer header). This is obfuscation, not
  // anonymization: hashSeed is a fast non-keyed hash and a low-entropy email
  // is still brute-forceable, so the result is best treated as a stable derived
  // identifier. The proper fix is a random per-user avatar seed persisted
  // server-side at signup (uncorrelated with email) — tracked as a follow-up.
  // The avatar style stays derived from `raw` so it doesn't shift for existing users.
  const key = hashSeed(raw).toString(36);
  const chosen = style || pickAvatarStyle(raw);
  return `https://api.dicebear.com/9.x/${chosen}/svg?seed=${encodeURIComponent(key)}`;
}
