/**
 * Theme utility helpers for the Dash app.
 * All functions are pure with no side-effects unless noted.
 */

export type ColorScheme = 'light' | 'dark' | 'system';

/** Return true when the system prefers a dark color scheme. */
export function systemPrefersDark(): boolean {
  if (typeof window === 'undefined') return false;
  return window.matchMedia('(prefers-color-scheme: dark)').matches;
}

/** Resolve the effective color scheme given a user preference and system preference. */
export function resolveColorScheme(preference: ColorScheme, systemDark: boolean): 'light' | 'dark' {
  if (preference === 'system') return systemDark ? 'dark' : 'light';
  return preference;
}

/** Return the CSS class to apply for the given color scheme. Pure. */
export function colorSchemeClass(scheme: 'light' | 'dark'): string {
  return scheme === 'dark' ? 'dark' : 'light';
}

/** Return true when the color scheme is dark mode. Pure. */
export function isDarkMode(scheme: ColorScheme, systemDark = false): boolean {
  return resolveColorScheme(scheme, systemDark) === 'dark';
}

/** Toggle between light and dark (ignores system). Pure. */
export function toggleScheme(current: ColorScheme): 'light' | 'dark' {
  if (current === 'dark') return 'light';
  return 'dark';
}

/** Return an ordered list of available color scheme options. Pure. */
export function colorSchemeOptions(): Array<{ value: ColorScheme; label: string }> {
  return [
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
    { value: 'system', label: 'System' },
  ];
}
