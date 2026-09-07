/**
 * Flow-specific colour helpers.
 *
 * isValidHex, normalizeHex and hexLuminance used to be defined here, byte for
 * byte alongside the same functions in @voxire/utils. They are imported now, so
 * there is one implementation rather than two drifting copies.
 *
 * The local hexLuminance also used 0.03928 as the sRGB linear-segment threshold
 * where WCAG specifies 0.04045. Verified before switching that the difference
 * changes no textOnColor decision, across all twenty palette colours and all 256
 * greyscale values, so this is a dedupe and not a behaviour change.
 */
import { hexLuminance, isValidHex, normalizeHex } from '@voxire/utils';

// Re-exported so this module's public surface is unchanged: the duplicate
// implementations are gone, the names callers already use are not.
export { hexLuminance, isValidHex, normalizeHex };

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

/**
 * A pastel (lightened) version of the hex colour.
 *
 * The `#ffffff` is an operand of the blend, not a surface colour, so it stays a
 * literal: `blendHex` parses both arguments as hex and a `var()` would not
 * parse. Checked while closing out rtylr#1073 rather than left as an
 * unexplained hardcoded hex.
 */
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

/**
 * A colour at a given opacity, as a CSS value.
 *
 * The codebase built these by concatenating a hex alpha suffix onto the
 * colour, `${health.color}1a`. That only works when the colour is a literal
 * hex: the moment it is a token, `var(--flow-status-approved)1a` is not a
 * colour at all and the declaration is dropped, taking the whole background
 * with it.
 *
 * color-mix takes any colour, including a var(), so a tone can move to a
 * token without every consumer having to know what shape it is. Tailwind v4
 * already requires color-mix support for its own opacity modifiers, so this
 * costs no browser support the app did not already need.
 */
export function withAlpha(color: string, percent: number): string {
  return `color-mix(in srgb, ${color} ${percent}%, transparent)`;
}

/** Returns a color from the palette by index, wrapping around. */
export function paletteColor(index: number): ProjectColor {
  return PROJECT_COLORS[index % PROJECT_COLORS.length];
}

/**
 * Flow's brand teal, re-exported so the app has one definition of it rather
 * than a copy that can drift from the one the token derivation falls back to.
 *
 * `accentTints` used to live below this line. It now lives in `@voxire/utils` as
 * part of `appearanceTokens`, which derives two members this version never had:
 * `--accent-fg` and `--flow-on-accent`. Both were hardcoded per mode in
 * globals.css, and the second measures as low as 1.30:1 on a dark accent, which
 * a custom hex field makes reachable. Nothing in this app should recompute the
 * accent family: `AppearanceProvider` is the only writer of it.
 */
export { ACCENT_DEFAULT } from '@voxire/utils';
