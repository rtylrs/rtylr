/**
 * Pure colour maths for the appearance layer.
 *
 * Nothing here touches the DOM, React, or a store. That is the point: the
 * accessibility claim in the design rests on these functions, and a pure
 * function is the only kind that can be proved rather than asserted. See
 * colorMath.test.ts for the sweep that does the proving.
 *
 * Every function tolerates malformed input instead of throwing. A colour
 * arrives here from a free-form hex field, so "unparseable" is a normal case,
 * not an exceptional one.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr)
 */

import { CONTRAST_TARGET } from './types';

export interface Rgb {
  /** 0-255 */
  r: number;
  /** 0-255 */
  g: number;
  /** 0-255 */
  b: number;
}

export interface Hsl {
  /** 0-360 */
  h: number;
  /** 0-100 */
  s: number;
  /** 0-100 */
  l: number;
}

/** Text on an accent fill, dark option. Matches the value globals.css hardcodes. */
export const NEAR_BLACK = '#0a0a0a';

/** Text on an accent fill, light option. */
export const NEAR_WHITE = '#ffffff';

/**
 * Mirrors apps/flow/src/shared/lib/colorUtils.ts so a value accepted there is
 * accepted here. Six digits only: the shorthand form is deliberately rejected,
 * because accepting it silently in one place and not another is worse than
 * rejecting it in both.
 */
export function isValidHex(color: string): boolean {
  return /^#?[0-9a-fA-F]{6}$/.test(color);
}

/** Normalizes a hex colour to lowercase with a leading #. */
export function normalizeHex(color: string): string {
  const hex = color.startsWith('#') ? color : `#${color}`;
  return hex.toLowerCase();
}

/**
 * Normalize when possible, fall back when not. This is the single place an
 * invalid colour becomes a valid one, so callers never have to branch on it.
 */
export function safeHex(color: string | null | undefined, fallback: string): string {
  if (typeof color === 'string' && isValidHex(color)) return normalizeHex(color);
  return normalizeHex(fallback);
}

/**
 * Parses a 6-digit hex. Returns black for anything unparseable rather than
 * throwing or returning NaN channels, which would poison every ratio computed
 * downstream and turn one bad input into an unreadable screen.
 */
export function hexToRgb(hex: string): Rgb {
  if (!isValidHex(hex)) return { r: 0, g: 0, b: 0 };
  const h = normalizeHex(hex).slice(1);
  return {
    r: parseInt(h.slice(0, 2), 16),
    g: parseInt(h.slice(2, 4), 16),
    b: parseInt(h.slice(4, 6), 16),
  };
}

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

function channelToHex(value: number): string {
  return clamp(Math.round(value), 0, 255).toString(16).padStart(2, '0');
}

export function rgbToHex({ r, g, b }: Rgb): string {
  return `#${channelToHex(r)}${channelToHex(g)}${channelToHex(b)}`;
}

export function rgbToHsl({ r, g, b }: Rgb): Hsl {
  const rn = clamp(r, 0, 255) / 255;
  const gn = clamp(g, 0, 255) / 255;
  const bn = clamp(b, 0, 255) / 255;
  const max = Math.max(rn, gn, bn);
  const min = Math.min(rn, gn, bn);
  const l = (max + min) / 2;
  const delta = max - min;

  if (delta === 0) return { h: 0, s: 0, l: l * 100 };

  const s = delta / (1 - Math.abs(2 * l - 1));
  let h: number;
  if (max === rn) h = ((gn - bn) / delta) % 6;
  else if (max === gn) h = (bn - rn) / delta + 2;
  else h = (rn - gn) / delta + 4;
  h *= 60;
  if (h < 0) h += 360;

  return { h, s: s * 100, l: l * 100 };
}

export function hslToRgb({ h, s, l }: Hsl): Rgb {
  const hn = ((h % 360) + 360) % 360;
  const sn = clamp(s, 0, 100) / 100;
  const ln = clamp(l, 0, 100) / 100;
  const c = (1 - Math.abs(2 * ln - 1)) * sn;
  const x = c * (1 - Math.abs(((hn / 60) % 2) - 1));
  const m = ln - c / 2;

  let rgb: [number, number, number];
  if (hn < 60) rgb = [c, x, 0];
  else if (hn < 120) rgb = [x, c, 0];
  else if (hn < 180) rgb = [0, c, x];
  else if (hn < 240) rgb = [0, x, c];
  else if (hn < 300) rgb = [x, 0, c];
  else rgb = [c, 0, x];

  return {
    r: (rgb[0] + m) * 255,
    g: (rgb[1] + m) * 255,
    b: (rgb[2] + m) * 255,
  };
}

export function hexToHsl(hex: string): Hsl {
  return rgbToHsl(hexToRgb(hex));
}

export function hslToHex(hsl: Hsl): string {
  return rgbToHex(hslToRgb(hsl));
}

/**
 * WCAG 2.x relative luminance.
 *
 * The linearization threshold is 0.04045 as published. Flow's older copy in
 * colorUtils.ts uses 0.03928 from the earlier erratum; for 8-bit sRGB input the
 * two are identical, because the band between them spans channel values 10.02
 * to 10.31 and contains no integer. So this is a correctness tidy-up with no
 * behavioural difference, not a change of results.
 */
export function relativeLuminance({ r, g, b }: Rgb): number {
  const toLinear = (channel: number): number => {
    const c = clamp(channel, 0, 255) / 255;
    return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
  };
  return 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b);
}

export function hexLuminance(hex: string): number {
  return relativeLuminance(hexToRgb(hex));
}

/** WCAG contrast ratio between two opaque colours. Always >= 1, order-free. */
export function contrastRatio(a: string, b: string): number {
  const la = hexLuminance(a);
  const lb = hexLuminance(b);
  const lighter = Math.max(la, lb);
  const darker = Math.min(la, lb);
  return (lighter + 0.05) / (darker + 0.05);
}

/**
 * The extreme that contrasts best with a ground. Also the direction the
 * lightness walk travels: away from the ground, never toward it.
 */
function bestExtreme(surface: string): { hex: string; darker: boolean } {
  const towardBlack = contrastRatio('#000000', surface);
  const towardWhite = contrastRatio('#ffffff', surface);
  return towardBlack >= towardWhite
    ? { hex: '#000000', darker: true }
    : { hex: '#ffffff', darker: false };
}

/** Lightness step, in HSL percentage points. */
const WALK_STEP = 0.5;

/**
 * The accent, restated at whatever lightness clears CONTRAST_TARGET against the
 * given surface.
 *
 * Hue and saturation are held fixed, so the result still reads as the colour the
 * user chose. That is the whole point: the token this supersedes was a hardcoded
 * teal that never failed contrast but sat up to 90 degrees off the fill beside
 * it, and "derive a safe grey" would trade one cohesion defect for another.
 *
 * An accent that already clears the target is returned unchanged, which is what
 * keeps Flow's shipped teal on dark exactly as it ships.
 *
 * HSL lightness is monotone in luminance for fixed hue and saturation, so a
 * one-directional walk cannot step over a passing value. If the walk exhausts
 * its bound the surface is a mid-tone that no tint of this hue can clear, and
 * the black-or-white extreme is returned instead.
 */
export function deriveAccentFg(accent: string, surface: string): string {
  const hex = safeHex(accent, '#000000');
  const ground = safeHex(surface, '#000000');

  if (contrastRatio(hex, ground) >= CONTRAST_TARGET) return hex;

  const { hex: extreme, darker } = bestExtreme(ground);
  const { h, s, l } = hexToHsl(hex);

  for (let step = WALK_STEP; step <= 100; step += WALK_STEP) {
    const nextL = darker ? l - step : l + step;
    if (nextL < 0 || nextL > 100) break;
    const candidate = hslToHex({ h, s, l: nextL });
    if (contrastRatio(candidate, ground) >= CONTRAST_TARGET) return candidate;
  }

  return extreme;
}

/**
 * Text sitting on the accent fill: near-black or white, whichever measures
 * better against that specific accent.
 *
 * Measured, never assumed. The token this replaces is hardcoded near-black,
 * which is safe for every curated swatch and every existing workspace colour,
 * and measures between 1.30:1 and 1.63:1 on the dark hexes a free-form picker
 * lets through. The assumption was correct right up to the moment the picker
 * shipped, which is exactly the kind of assumption to delete.
 */
export function deriveOnAccent(accent: string): string {
  const hex = safeHex(accent, '#000000');
  return contrastRatio(NEAR_BLACK, hex) >= contrastRatio(NEAR_WHITE, hex) ? NEAR_BLACK : NEAR_WHITE;
}

/**
 * White text measuring at least this on an accent marks it a mid-tone rather
 * than a pastel. 3:1 is WCAG's bar for large text and UI components.
 */
const MID_TONE_WHITE = 3;

/**
 * The fill for a solid button in this accent; put deriveOnAccent(fill) on it.
 *
 * On a saturated mid-tone deriveOnAccent alone picks near-black by a hair:
 * violet #8b5cf6 measures 4.7:1 with near-black against 4.2:1 with white. That
 * passes AA and still looks wrong, because nobody ships black text on a violet
 * button. So when white already gets 3:1, the fill walks darker, hue and
 * saturation held, until white clears CONTRAST_TARGET. Light accents (white
 * under 3:1) and dark ones (white already clear) come back unchanged.
 */
export function deriveActionFill(accent: string): string {
  const hex = safeHex(accent, '#000000');
  const white = contrastRatio(NEAR_WHITE, hex);
  if (white < MID_TONE_WHITE || white >= CONTRAST_TARGET) return hex;

  const { h, s, l } = hexToHsl(hex);
  for (let nextL = l - WALK_STEP; nextL >= 0; nextL -= WALK_STEP) {
    const candidate = hslToHex({ h, s, l: nextL });
    if (contrastRatio(NEAR_WHITE, candidate) >= CONTRAST_TARGET) return candidate;
  }
  return hex;
}

/** `rgba(r, g, b, a)` from a hex plus an alpha. Matches the accentTints output. */
export function rgbaFromHex(hex: string, alpha: number): string {
  const { r, g, b } = hexToRgb(hex);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/** `rgba(r, g, b, a)` from an "r,g,b" triple, as PalettePreset.textRgb stores it. */
export function rgbaFromTriple(triple: string, alpha: number): string {
  const parts = triple.split(',').map((p) => Number(p.trim()));
  const [r, g, b] = parts.length === 3 && parts.every(Number.isFinite) ? parts : [0, 0, 0];
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

/**
 * The opaque colour an alpha fill actually resolves to over a given ground.
 *
 * Needed because a contrast ratio is only meaningful between two opaque
 * colours. `--accent-muted` is `rgba(accent, 0.12..0.18)`, so the colour a
 * reader's eye receives is the accent composited over whatever sits behind it,
 * never the accent itself. Deriving text against the raw accent, or against the
 * ground while ignoring the fill on top of it, both overstate the contrast.
 */
export function compositeOver(hex: string, alpha: number, ground: string): string {
  const fg = hexToRgb(safeHex(hex, '#000000'));
  const bg = hexToRgb(safeHex(ground, '#000000'));
  const a = Math.min(1, Math.max(0, alpha));
  const mix = (f: number, b: number): number => Math.round(a * f + (1 - a) * b);
  return rgbToHex({ r: mix(fg.r, bg.r), g: mix(fg.g, bg.g), b: mix(fg.b, bg.b) });
}

/**
 * A hue-preserving accent that clears the contrast target against EVERY ground
 * it might be laid over, not just one.
 *
 * `deriveAccentFg` guarantees its result against the single surface it is given.
 * That guarantee does not transfer: `--accent-fg` was derived against
 * `palette.card` and then used on `--accent-muted`, where 40 of 56 light-mode
 * palette/accent combinations measured below 4.5:1 (rtylr#970).
 *
 * Rather than pick the "right" ground and be wrong again the next time a
 * consumer uses the token somewhere new, derive a candidate for each ground and
 * keep whichever holds up worst-case across all of them.
 */
export function deriveAccentFgAcross(accent: string, grounds: readonly string[]): string {
  const list = grounds.filter((g) => isValidHex(g));
  if (list.length === 0) return deriveAccentFg(accent, NEAR_WHITE);

  const worstCase = (candidate: string): number =>
    Math.min(...list.map((g) => contrastRatio(candidate, g)));

  let best = deriveAccentFg(accent, list[0]);
  let bestScore = worstCase(best);
  for (const ground of list.slice(1)) {
    const candidate = deriveAccentFg(accent, ground);
    const score = worstCase(candidate);
    if (score > bestScore) {
      best = candidate;
      bestScore = score;
    }
  }
  return best;
}
