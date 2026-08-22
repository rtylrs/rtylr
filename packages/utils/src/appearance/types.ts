/**
 * Shared shapes for the appearance layer.
 *
 * This file is the contract three otherwise independent pieces agree on: the
 * preferences store, the pure token derivation, and the provider that writes the
 * result to the DOM. Nothing here has behaviour, so it can be read on its own.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr)
 */

export type ThemeMode = 'light' | 'dark' | 'system';

/** What `mode` resolves to once `system` has been read from the media query. */
export type ResolvedMode = 'light' | 'dark';

export type Density = 'compact' | 'cozy' | 'roomy';
export type TextSize = 's' | 'm' | 'l';
export type Corners = 'sharp' | 'soft' | 'round';

export interface AppearancePrefs {
  mode: ThemeMode;
  /** Preset id, resolved within the active mode's family. */
  palette: string;
  /**
   * Hex accent, or null to follow the active workspace colour.
   *
   * The null is load-bearing: it expresses "user accent overrides workspace"
   * without storing a copy of the workspace colour that would then go stale, and
   * it is what the reset control writes.
   */
  accent: string | null;
  density: Density;
  textSize: TextSize;
  corners: Corners;
  reduceMotion: boolean;
}

/** A named ground. Presets change surfaces only; they never touch the accent. */
export interface PalettePreset {
  id: string;
  /** i18n key, not a display string. Palette names are user-facing copy. */
  labelKey: string;
  mode: ResolvedMode;
  bg: string;
  surface: string;
  card: string;
  /** Base text channel as "r,g,b". Alphas are applied per mode by the derivation. */
  textRgb: string;
}

export interface AccentSwatch {
  id: string;
  /** i18n key, not a display string. */
  labelKey: string;
  hex: string;
}

/**
 * CSS custom property name to value. Deliberately a flat record: the provider
 * writes it with setProperty in a loop and never needs to know the key set.
 */
export type AppearanceTokens = Record<string, string>;

/**
 * Inputs the derivation needs beyond the preferences themselves.
 *
 * `workspaceAccent` is the fallback used when `prefs.accent` is null. It is
 * passed in rather than read from a store so the derivation stays pure and
 * testable without a Flow workspace in scope.
 */
export interface AppearanceContext {
  resolvedMode: ResolvedMode;
  workspaceAccent: string;
}

/**
 * Minimum contrast the derived foregrounds must clear, in a WCAG ratio.
 *
 * 4.6 rather than 4.5 on purpose. Stopping at the first value that merely
 * reaches 4.5 produced exactly 4.50:1 on every light palette, which leaves
 * nothing for rounding or for a later change to a surface token. The extra
 * tenth costs an imperceptible darkening. Flow's type floor is 10px, so WCAG's
 * large-text allowance never applies and 4.5 is the only bar that counts.
 */
export const CONTRAST_TARGET = 4.6;

/** The bar the test suite asserts against, as opposed to the one we aim for. */
export const CONTRAST_FLOOR = 4.5;
