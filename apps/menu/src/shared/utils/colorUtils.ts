/**
 * Color utility helpers for the Menu app.
 */

export const MENU_COLORS = {
  primary: '#0F172A',
  accent: '#F97316',
  success: '#22C55E',
  warning: '#EAB308',
  error: '#EF4444',
  muted: '#6B7280',
} as const;

export type MenuColorKey = keyof typeof MENU_COLORS;

/** Return the hex color value for a known menu color key. */
export function getMenuColor(key: MenuColorKey): string {
  return MENU_COLORS[key];
}

/**
 * Convert a hex color to an rgba string.
 * Accepts 3 or 6-digit hex with or without the '#' prefix.
 */
export function hexToRgba(hex: string, alpha = 1): string {
  const clean = hex.startsWith('#') ? hex.slice(1) : hex;
  const expanded =
    clean.length === 3
      ? clean
          .split('')
          .map((c) => c + c)
          .join('')
      : clean;
  const r = parseInt(expanded.slice(0, 2), 16);
  const g = parseInt(expanded.slice(2, 4), 16);
  const b = parseInt(expanded.slice(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/** Return true if the value is a valid 3 or 6-digit hex color starting with '#'. */
export function isValidHex(value: string): boolean {
  return /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/.test(value);
}
