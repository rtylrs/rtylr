export const ASSETS = {
  logos: {
    dark: 'https://cdn.voxire.dev/web_content/rtylr/logo-dark.webp',
    light: 'https://cdn.voxire.dev/web_content/rtylr/logo-light.webp',
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
