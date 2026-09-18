/**
 * The palette and accent data.
 *
 * Every value here was measured, not chosen by eye. The full matrix (8 palettes
 * times 14 accent fixtures) was verified to clear 4.5:1 on both derived
 * foregrounds, worst case 4.60:1. If you add a palette or a swatch, the contrast
 * sweep in appearanceTokens.test.ts covers it automatically. If it fails, the
 * value is wrong, not the test.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr)
 */

import type { AccentSwatch, PalettePreset } from './types';

/** Flow's shipped teal. Also the fallback when a workspace has no colour. */
export const ACCENT_DEFAULT = '#64c5b9';

/**
 * Four per mode, so the card grid balances in both. The dark values for
 * `midnight` and the light values for `paper` are Flow as it ships today, lifted
 * from apps/flow/src/globals.css: changing them would re-theme the app rather
 * than add a choice to it.
 */
export const PALETTES: readonly PalettePreset[] = [
  {
    id: 'midnight',
    labelKey: 'appearance.palettes.midnight',
    mode: 'dark',
    bg: '#0a0a0a',
    surface: '#111111',
    card: '#161616',
    textRgb: '240,240,238',
  },
  {
    id: 'graphite',
    labelKey: 'appearance.palettes.graphite',
    mode: 'dark',
    bg: '#131316',
    surface: '#1a1a20',
    card: '#202027',
    textRgb: '238,238,242',
  },
  {
    id: 'ink',
    labelKey: 'appearance.palettes.ink',
    mode: 'dark',
    bg: '#080b12',
    surface: '#0d1119',
    card: '#121722',
    textRgb: '232,238,248',
  },
  {
    id: 'moss',
    labelKey: 'appearance.palettes.moss',
    mode: 'dark',
    bg: '#0a0d0b',
    surface: '#101511',
    card: '#161d18',
    textRgb: '234,241,235',
  },
  {
    id: 'paper',
    labelKey: 'appearance.palettes.paper',
    mode: 'light',
    bg: '#f4f4f5',
    surface: '#ffffff',
    card: '#ffffff',
    textRgb: '9,9,11',
  },
  {
    id: 'linen',
    labelKey: 'appearance.palettes.linen',
    mode: 'light',
    bg: '#f7f5f1',
    surface: '#fffefb',
    card: '#fffefb',
    textRgb: '26,22,16',
  },
  {
    id: 'mist',
    labelKey: 'appearance.palettes.mist',
    mode: 'light',
    // #e2e9f4 rather than #edf0f4. At the lighter value mist sat 2.41 dE from
    // paper, which is barely past the just-noticeable threshold, so the two read
    // as the same choice in a 56px card. This raises the minimum pairwise
    // separation across the four light palettes from 2.41 to 2.67 and moves the
    // binding constraint to paper against linen, which is inherent: a warm and a
    // neutral near-white cannot be pulled much further apart while both stay
    // light. Text contrast is 15.48:1, so there is no accessibility cost.
    bg: '#e2e9f4',
    surface: '#ffffff',
    card: '#ffffff',
    textRgb: '12,17,26',
  },
  {
    id: 'slate',
    labelKey: 'appearance.palettes.slate',
    mode: 'light',
    bg: '#e5e9ef',
    surface: '#f7f9fc',
    card: '#ffffff',
    textRgb: '12,17,26',
  },
] as const;

/**
 * Ten curated accents. All are light enough that near-black text sits safely on
 * them, which is why a curated-only product would have no contrast problem at
 * all. The risk arrives with the custom hex field, not with this list.
 */
export const ACCENT_SWATCHES: readonly AccentSwatch[] = [
  { id: 'teal', labelKey: 'appearance.accents.teal', hex: '#64c5b9' },
  { id: 'iris', labelKey: 'appearance.accents.iris', hex: '#8b8cf5' },
  { id: 'sky', labelKey: 'appearance.accents.sky', hex: '#56b6e8' },
  { id: 'mint', labelKey: 'appearance.accents.mint', hex: '#5ed6a0' },
  { id: 'lime', labelKey: 'appearance.accents.lime', hex: '#a3d84b' },
  { id: 'amber', labelKey: 'appearance.accents.amber', hex: '#e0a23c' },
  { id: 'coral', labelKey: 'appearance.accents.coral', hex: '#f0805a' },
  { id: 'rose', labelKey: 'appearance.accents.rose', hex: '#f56d8f' },
  { id: 'violet', labelKey: 'appearance.accents.violet', hex: '#b57ce8' },
  { id: 'slate', labelKey: 'appearance.accents.slate', hex: '#8f9bab' },
] as const;

/**
 * Dark accents used only as test fixtures, never offered in the UI.
 *
 * These are what a free-form colour picker lets through, and they are the only
 * values that exercise the --flow-on-accent failure at all: hardcoded near-black
 * measures between 1.30:1 and 1.63:1 against them. A sweep over the curated
 * swatches alone would pass while proving nothing, so the suite needs these.
 */
export const STRESS_ACCENTS: readonly string[] = [
  '#1b2a5c',
  '#2d1b4e',
  '#0f3d2e',
  '#5c1b1b',
] as const;

/** Presets available for a mode. The UI never offers a light preset in dark. */
export function palettesForMode(mode: 'light' | 'dark'): readonly PalettePreset[] {
  return PALETTES.filter((p) => p.mode === mode);
}

/**
 * Resolve a stored palette id against a mode, falling back to that mode's first
 * preset. A stored id can legitimately belong to the other family: switching
 * mode does not rewrite the stored palette, so this has to cope rather than
 * throw.
 */
export function resolvePalette(id: string, mode: 'light' | 'dark'): PalettePreset {
  const family = palettesForMode(mode);
  return family.find((p) => p.id === id) ?? family[0];
}
