export type ErpTheme = 'dark' | 'light' | 'system';

export const THEME_STORAGE_KEY = 'rtylr-erp-theme';

/** Applies the theme by setting `data-theme` on the document root. */
export function applyTheme(theme: ErpTheme): void {
  const root = document.documentElement;
  if (theme === 'system') {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    root.setAttribute('data-theme', prefersDark ? 'dark' : 'light');
  } else {
    root.setAttribute('data-theme', theme);
  }
}

/** Reads the persisted theme, falling back to dark (the app default). */
export function readStoredTheme(): ErpTheme {
  const stored = localStorage.getItem(THEME_STORAGE_KEY);
  return stored === 'light' || stored === 'system' ? stored : 'dark';
}
