/**
 * The appearance layer's public surface.
 *
 * Three pieces that deliberately do not know about each other: a pure
 * preferences store, a pure token derivation, and the colour maths the
 * derivation is built from. The DOM writer lives in `@voxire/ui` and is the only
 * thing here that touches `documentElement`.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr)
 */

export type {
  AccentSwatch,
  AppearanceContext,
  AppearancePrefs,
  AppearanceTokens,
  Corners,
  Density,
  PalettePreset,
  ResolvedMode,
  TextSize,
  ThemeMode,
} from './types';
export { CONTRAST_FLOOR, CONTRAST_TARGET } from './types';

export {
  ACCENT_DEFAULT,
  ACCENT_SWATCHES,
  PALETTES,
  STRESS_ACCENTS,
  palettesForMode,
  resolvePalette,
} from './palettes';

export type {
  AppearanceActions,
  AppearanceStore,
  AppearanceStoreHook,
  CreateAppearanceStoreOptions,
} from './appearanceStore';
export {
  APPEARANCE_DEFAULTS,
  LEGACY_THEME_STORAGE_KEY,
  createAppearanceStore,
  readLegacyMode,
  resolveMode,
} from './appearanceStore';

export {
  contrastRatio,
  compositeOver,
  deriveAccentFg,
  deriveAccentFgAcross,
  deriveOnAccent,
  hexLuminance,
  isValidHex,
  normalizeHex,
  relativeLuminance,
} from './colorMath';

export { appearanceTokens } from './appearanceTokens';
