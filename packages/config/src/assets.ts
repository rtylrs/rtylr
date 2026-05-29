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

export const AVATAR_STYLES = ['thumbs', 'bottts-neutral', 'fun-emoji', 'lorelei-neutral'] as const;
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
  const key = seed || 'user';
  const chosen = style || pickAvatarStyle(key);
  return `https://api.dicebear.com/9.x/${chosen}/svg?seed=${encodeURIComponent(key)}`;
}
