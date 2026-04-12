export const ASSETS = {
  logos: {
    dark: 'https://cdn.voxire.dev/web_content/rtylr/logo-dark.webp',
    light: 'https://cdn.voxire.dev/web_content/rtylr/logo-light.webp',
  },
  /** Per-app logos — will be populated when individual logos are designed.
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
