/**
 * The token derivation: preferences in, CSS custom properties out.
 *
 * Pure. No DOM, no React, no store import. The design picked this shape for one
 * reason: a pure function is the only kind whose accessibility claim can be
 * proved by a test rather than asserted in a comment. Everything about writing
 * these values to an element belongs to the provider.
 *
 * The surface and accent values here reproduce apps/flow/src/globals.css and
 * apps/flow/src/shared/lib/colorUtils.ts `accentTints`, alpha for alpha. The two
 * genuinely new tokens are --accent-fg and --flow-on-accent, which stop being
 * hardcoded per mode and start being derived from the accent in play.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr)
 */

import {
  compositeOver,
  deriveAccentFg,
  deriveAccentFgAcross,
  deriveOnAccent,
  rgbaFromHex,
  rgbaFromTriple,
  safeHex,
} from './colorMath';
import { ACCENT_DEFAULT, resolvePalette } from './palettes';
import type {
  AppearanceContext,
  AppearancePrefs,
  AppearanceTokens,
  Corners,
  Density,
  ResolvedMode,
  TextSize,
} from './types';

/**
 * Per-mode alphas, read off apps/flow/src/globals.css rather than remembered.
 *
 * The muted values are the interesting ones and both carry a comment in that
 * file: .5 measured 4.39:1 on --flow-card in dark and 3.74:1 on white in light,
 * so dark runs .58 and light runs .62. Light text is fully opaque because
 * globals.css declares it as `rgb(9, 9, 11)` with no alpha at all.
 *
 * `dim` never carries text in either mode. It is rules, dividers, and icon
 * strokes that repeat an adjacent label, which is why it is allowed to sit far
 * below the contrast floor.
 */
const TEXT_ALPHAS: Record<ResolvedMode, { text: number; muted: number; dim: number }> = {
  dark: { text: 0.92, muted: 0.58, dim: 0.28 },
  light: { text: 1, muted: 0.62, dim: 0.35 },
};

/**
 * Washes over the palette ground. White over dark, black over light, exactly as
 * globals.css does it: these are neutral overlays, not tints of the text colour.
 */
const WASHES: Record<ResolvedMode, { overlay: string; border: number; hover: number }> = {
  dark: { overlay: '255, 255, 255', border: 0.07, hover: 0.05 },
  light: { overlay: '0, 0, 0', border: 0.08, hover: 0.05 },
};

/**
 * Accent wash alphas, lifted from `accentTints`.
 *
 * A .12 wash reads lighter on white than on black, so light mode runs the tint
 * and border heavier to carry the same weight. The focus alpha is deliberately
 * the same in both: it is already strong enough either way.
 */
const ACCENT_ALPHAS: Record<ResolvedMode, { muted: number; border: number; focus: number }> = {
  dark: { muted: 0.12, border: 0.3, focus: 0.5 },
  light: { muted: 0.18, border: 0.42, focus: 0.5 },
};

/**
 * The density multiplier, deliberately UNITLESS rather than a px length.
 *
 * A consumer multiplies its own rem base by this: `calc(0.25rem * var(--flow-u))`.
 * Emitting `3px` here instead would work, and it would also silently break
 * browser font-size scaling: a px length cannot be turned back into rem inside
 * calc(), so every spacing utility in the app would stop responding to a user's
 * default font size. A multiplier keeps the base in rem and still moves density.
 *
 * `1` is Tailwind's own step and therefore today's app. Compact is 0.75, so
 * anything relying on a 44px minimum has to be pinned to a literal rather than
 * left to scale: at 0.75 an `h-11` becomes 33px.
 */
const DENSITY_UNIT: Record<Density, string> = {
  compact: '0.75',
  cozy: '1',
  roomy: '1.25',
};

/**
 * The text-size multiplier, unitless for the same reason as density above.
 *
 * Nothing consumes this yet and nothing should until Flow's type stops being px
 * literals: 618 of its 637 font-size declarations are values like `text-[12px]`,
 * which no token can reach. Section 7 of the design has the detail, and the
 * control ships disabled rather than pretending otherwise.
 */
const TEXT_SIZE: Record<TextSize, string> = {
  s: '0.92',
  m: '1',
  l: '1.1',
};

/** Corner radius. `soft` is 8px, matching --radius-md in the shared preset. */
const CORNER_RADIUS: Record<Corners, string> = {
  sharp: '4px',
  soft: '8px',
  round: '14px',
};

/**
 * A scale lookup that cannot come back undefined.
 *
 * The preference types make an out-of-range value impossible at compile time,
 * but these values arrive from a persisted store, and a stale or hand-edited
 * localStorage entry is not bound by the type. Emitting `undefined` into a
 * custom property would drop the declaration silently.
 */
function pick<K extends string>(scale: Record<K, string>, key: K, fallback: K): string {
  return scale[key] ?? scale[fallback];
}

/**
 * Derives every custom property the appearance layer owns.
 *
 * The accent source is `prefs.accent ?? ctx.workspaceAccent`, which is what
 * "user accent overrides workspace" means in practice: the null in
 * AppearancePrefs is the whole mechanism, and it is what a reset control writes.
 * An unparseable value at either position falls through to ACCENT_DEFAULT, so
 * this never throws and never emits a colour a browser would drop.
 *
 * `prefs.reduceMotion` is intentionally absent from the output. It is not a
 * colour or a length, so it belongs on an attribute the stylesheet can select
 * on, which is the provider's business rather than this function's.
 */
export function appearanceTokens(prefs: AppearancePrefs, ctx: AppearanceContext): AppearanceTokens {
  const mode = ctx.resolvedMode;
  const palette = resolvePalette(prefs.palette, mode);

  const accent = safeHex(prefs.accent ?? ctx.workspaceAccent, ACCENT_DEFAULT);
  const textAlpha = TEXT_ALPHAS[mode];
  const wash = WASHES[mode];
  const accentAlpha = ACCENT_ALPHAS[mode];

  return {
    // ── Ground ──
    '--flow-bg': palette.bg,
    '--flow-surface': palette.surface,
    '--flow-card': palette.card,
    '--flow-border': `rgba(${wash.overlay}, ${wash.border})`,
    '--flow-hover': `rgba(${wash.overlay}, ${wash.hover})`,
    // Dark floats the field above the surface with a wash; light declares the
    // surface itself, which is what globals.css does with its literal #ffffff.
    '--flow-input-bg': mode === 'dark' ? `rgba(${wash.overlay}, 0.04)` : palette.surface,

    // ── Text ──
    '--flow-text': rgbaFromTriple(palette.textRgb, textAlpha.text),
    '--flow-text-muted': rgbaFromTriple(palette.textRgb, textAlpha.muted),
    '--flow-text-dim': rgbaFromTriple(palette.textRgb, textAlpha.dim),

    // ── Accent family ──
    '--accent': accent,
    '--accent-muted': rgbaFromHex(accent, accentAlpha.muted),
    '--accent-border': rgbaFromHex(accent, accentAlpha.border),
    '--accent-focus': rgbaFromHex(accent, accentAlpha.focus),

    // ── Derived, and the reason this function exists ──
    // Measured against the card the text actually sits on, not against the
    // page background, because a chip label is on --flow-card.
    '--accent-fg': deriveAccentFg(accent, palette.card),
    '--flow-on-accent': deriveOnAccent(accent),
    // Text on the --accent-muted tint, which is a different background from the
    // card and therefore needs its own derivation. Reusing --accent-fg here put
    // 40 of the 56 light-mode palette/accent combinations below 4.5:1
    // (rtylr#970): a tint darkens the ground under already-dark text, and
    // --accent-fg has no headroom to spend because its walk stops the moment it
    // clears the target.
    //
    // Derived against the tint composited over all three grounds it can sit on
    // rather than a chosen one. Picking the ground is what went wrong the first
    // time, so this token does not depend on picking it right.
    '--accent-on-muted': deriveAccentFgAcross(
      accent,
      [palette.bg, palette.surface, palette.card].map((ground) =>
        compositeOver(accent, accentAlpha.muted, ground)
      )
    ),
    // The focus ring was the third hardcoded token, and it was missed twice:
    // once when --accent-fg and --flow-on-accent were derived, and again when a
    // browser check found a teal ring drawn around a coral-accented card. Same
    // defect, same fix. Derived against --flow-bg rather than the card, because
    // the ring sits at a 2px offset outside the control, on the page ground.
    '--flow-focus-ring': rgbaFromHex(deriveAccentFg(accent, palette.bg), 0.9),

    // ── Shape ──
    '--flow-u': pick(DENSITY_UNIT, prefs.density, 'cozy'),
    '--flow-fs': pick(TEXT_SIZE, prefs.textSize, 'm'),
    '--flow-r': pick(CORNER_RADIUS, prefs.corners, 'soft'),
  };
}
