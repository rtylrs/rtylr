/**
 * Color utility helpers for the Dash UI.
 * All functions are pure with no side-effects.
 */

/** Convert a hex color to an rgba string. Accepts 3 or 6-digit hex with or without '#'. */
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

/** Lighten a hex color by mixing it with white by the given percentage (0-100). */
export function lighten(hex: string, percent: number): string {
  const clean = (hex.startsWith('#') ? hex.slice(1) : hex).padEnd(6, '0');
  const r = parseInt(clean.slice(0, 2), 16);
  const g = parseInt(clean.slice(2, 4), 16);
  const b = parseInt(clean.slice(4, 6), 16);
  const factor = Math.min(1, Math.max(0, percent / 100));
  const mix = (channel: number) => Math.round(channel + (255 - channel) * factor);
  const toHex = (n: number) => n.toString(16).padStart(2, '0');
  return `#${toHex(mix(r))}${toHex(mix(g))}${toHex(mix(b))}`;
}

/** Darken a hex color by mixing it with black by the given percentage (0-100). */
export function darken(hex: string, percent: number): string {
  const clean = (hex.startsWith('#') ? hex.slice(1) : hex).padEnd(6, '0');
  const r = parseInt(clean.slice(0, 2), 16);
  const g = parseInt(clean.slice(2, 4), 16);
  const b = parseInt(clean.slice(4, 6), 16);
  const factor = Math.min(1, Math.max(0, percent / 100));
  const mix = (channel: number) => Math.round(channel * (1 - factor));
  const toHex = (n: number) => n.toString(16).padStart(2, '0');
  return `#${toHex(mix(r))}${toHex(mix(g))}${toHex(mix(b))}`;
}

/** Generate a deterministic color from a string (useful for avatars/tags). */
export function stringToColor(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = str.charCodeAt(i) + ((hash << 5) - hash);
    hash |= 0;
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 65%, 50%)`;
}

/** Return a contrast color (black or white) for a given hex background. */
export function contrastColor(hex: string): '#000000' | '#ffffff' {
  const clean = (hex.startsWith('#') ? hex.slice(1) : hex).padEnd(6, '0');
  const r = parseInt(clean.slice(0, 2), 16);
  const g = parseInt(clean.slice(2, 4), 16);
  const b = parseInt(clean.slice(4, 6), 16);
  // Relative luminance per WCAG formula
  const luminance = (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
  return luminance > 0.5 ? '#000000' : '#ffffff';
}
