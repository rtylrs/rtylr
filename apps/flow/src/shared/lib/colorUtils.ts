/** Returns true if the string is a valid 6-digit hex color (with or without #). */
export function isValidHex(color: string): boolean {
  return /^#?[0-9a-fA-F]{6}$/.test(color);
}

/** Normalizes a hex color to lowercase with leading #. */
export function normalizeHex(color: string): string {
  const hex = color.startsWith('#') ? color : `#${color}`;
  return hex.toLowerCase();
}

/** Returns the luminance of a hex color (0-1). */
export function hexLuminance(hex: string): number {
  const clean = hex.replace('#', '');
  const r = parseInt(clean.slice(0, 2), 16) / 255;
  const g = parseInt(clean.slice(2, 4), 16) / 255;
  const b = parseInt(clean.slice(4, 6), 16) / 255;
  const toLinear = (c: number) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4);
  return 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b);
}

/** Returns 'dark' or 'light' based on whether the background color is dark. */
export function textOnColor(hex: string): 'dark' | 'light' {
  return hexLuminance(hex) > 0.179 ? 'dark' : 'light';
}

/** Blends two hex colors at a given ratio (0 = fully a, 1 = fully b). */
export function blendHex(a: string, b: string, ratio: number): string {
  const parse = (h: string) => [
    parseInt(h.slice(1, 3), 16),
    parseInt(h.slice(3, 5), 16),
    parseInt(h.slice(5, 7), 16),
  ];
  const [ar, ag, ab] = parse(normalizeHex(a));
  const [br, bg, bb] = parse(normalizeHex(b));
  const r = Math.round(ar + (br - ar) * ratio);
  const g = Math.round(ag + (bg - ag) * ratio);
  const bl = Math.round(ab + (bb - ab) * ratio);
  return `#${[r, g, bl].map((v) => v.toString(16).padStart(2, '0')).join('')}`;
}

/** Returns a pastel (lightened) version of the hex color. */
export function toPastel(hex: string): string {
  return blendHex(hex, '#ffffff', 0.6);
}

/** Predefined project color palette. */
export const PROJECT_COLORS = [
  '#3b82f6',
  '#8b5cf6',
  '#ec4899',
  '#ef4444',
  '#f97316',
  '#eab308',
  '#22c55e',
  '#14b8a6',
  '#06b6d4',
  '#6366f1',
] as const;

export type ProjectColor = (typeof PROJECT_COLORS)[number];

/** Returns a color from the palette by index, wrapping around. */
export function paletteColor(index: number): ProjectColor {
  return PROJECT_COLORS[index % PROJECT_COLORS.length];
}

/** Resolved theme, after 'system' has been collapsed to one or the other. */
export type ResolvedTheme = 'dark' | 'light';

const ACCENT_FALLBACK = '#64c5b9';

/**
 * The four accent custom properties, derived from the active workspace colour.
 *
 * These have to be computed at runtime because the accent is per-workspace, and
 * they are applied as inline styles on documentElement - which beats any
 * stylesheet rule. That is why the per-mode alpha lives here rather than in
 * globals.css: a light-mode value declared there would be silently overridden.
 *
 * A .12 wash reads lighter on white than on black, so light mode runs the tint
 * and border heavier to carry the same weight. The focus alpha is unchanged: it
 * is already strong enough in both modes.
 */
export function accentTints(color: string, theme: ResolvedTheme): Record<string, string> {
  const hex = isValidHex(color) ? normalizeHex(color) : ACCENT_FALLBACK;
  const h = hex.replace('#', '');
  const r = parseInt(h.slice(0, 2), 16);
  const g = parseInt(h.slice(2, 4), 16);
  const b = parseInt(h.slice(4, 6), 16);
  const rgba = (alpha: number): string => `rgba(${r}, ${g}, ${b}, ${alpha})`;
  const light = theme === 'light';
  return {
    '--accent': hex,
    '--accent-muted': rgba(light ? 0.18 : 0.12),
    '--accent-border': rgba(light ? 0.42 : 0.3),
    '--accent-focus': rgba(0.5),
  };
}
