/**
 * The appearance preferences store.
 *
 * Pure state, deliberately. This file never touches the DOM: no setProperty, no
 * setAttribute, no matchMedia. The provider in `@voxire/ui` is the single writer,
 * and that is not a stylistic preference. An inline custom property on
 * documentElement outranks every stylesheet rule, so a second writer here would
 * silently defeat `globals.css` in whichever app adopted it. `resolveMode` takes
 * `prefersDark` as an argument for the same reason: the media query is the
 * caller's to read.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr),
 * section 4.1.
 */

import { create } from 'zustand';
import type { Mutate, StoreApi, UseBoundStore } from 'zustand';
import { persist } from 'zustand/middleware';

import type { AppearancePrefs, Corners, Density, ResolvedMode, TextSize, ThemeMode } from './types';

/**
 * The shipped defaults.
 *
 * `mode` is dark because that is the mode Flow is designed in: every screen in
 * the design set is drawn dark, and a bare `:root` in globals.css resolves to
 * the dark palette. Light stays a first-class opt-in, and a stored preference
 * always wins over this.
 *
 * `accent` is null, meaning "follow the active workspace colour". Storing a copy
 * of the workspace colour instead would go stale the moment the workspace
 * changed, and null is also what the reset control writes.
 *
 * `reduceMotion` is false here rather than reading `prefers-reduced-motion`,
 * because reading it would be a DOM access. A caller that wants the setting to
 * default on when the system asks for it passes that in:
 * `createAppearanceStore(key, { reduceMotion: prefersReducedMotion })`.
 */
export const APPEARANCE_DEFAULTS: AppearancePrefs = {
  mode: 'dark',
  // 'midnight' is Flow as it ships today, lifted from globals.css. See palettes.ts.
  palette: 'midnight',
  accent: null,
  density: 'cozy',
  textSize: 'm',
  corners: 'soft',
  reduceMotion: false,
};

/**
 * The key Flow's hand-rolled themeStore persisted under, shape
 * `{ state: { theme: 'dark' | 'light' | 'system' } }`.
 *
 * Exported so the string is written down exactly once, but the package never
 * reads it on its own: Flow passes it as `legacyModeKey` at its own call site.
 * See `CreateAppearanceStoreOptions.legacyModeKey` for why it is opt-in.
 */
export const LEGACY_THEME_STORAGE_KEY = 'flow-theme';

/** Actions, one per field plus a reset. */
export interface AppearanceActions {
  setMode: (mode: ThemeMode) => void;
  setPalette: (palette: string) => void;
  setAccent: (accent: string | null) => void;
  setDensity: (density: Density) => void;
  setTextSize: (textSize: TextSize) => void;
  setCorners: (corners: Corners) => void;
  setReduceMotion: (reduceMotion: boolean) => void;
  /**
   * Back to the defaults this store was created with. A migrated legacy mode is
   * a stored choice, not a default, so reset discards it too.
   */
  reset: () => void;
}

export type AppearanceStore = AppearancePrefs & AppearanceActions;

/**
 * Factory options: the preference defaults, plus the one thing that is about
 * migration rather than about a preference.
 */
export interface CreateAppearanceStoreOptions extends Partial<AppearancePrefs> {
  /**
   * localStorage key of a pre-existing theme store to carry the mode across
   * from, shape `{ state: { theme } }`. Omit it and no legacy read happens at
   * all.
   *
   * Opt-in rather than baked in, because the legacy key belongs to one app.
   * Flow passes `LEGACY_THEME_STORAGE_KEY` so anyone who already chose light
   * keeps light. dash and menu are scheduled to adopt this factory later (spec
   * section 9) under their own keys, and a hardcoded `flow-theme` read would
   * hand each of them Flow's mode on first init. That bug happens exactly once
   * per install and then looks indistinguishable from a user preference, which
   * is what makes it expensive to trace rather than merely wrong.
   */
  legacyModeKey?: string | null;
}

/**
 * Only the preferences are persisted, so the stored shape is a readable
 * contract rather than whatever survived JSON.stringify.
 */
export type AppearanceStoreHook = UseBoundStore<
  Mutate<StoreApi<AppearanceStore>, [['zustand/persist', AppearancePrefs]]>
>;

/**
 * Resolve `system` against a system preference the caller supplies.
 *
 * Pure on purpose: the provider owns the media query and its change listener, so
 * this stays testable for both preferences without a DOM in scope.
 */
export function resolveMode(mode: ThemeMode, prefersDark: boolean): ResolvedMode {
  if (mode === 'system') return prefersDark ? 'dark' : 'light';
  return mode;
}

/**
 * The gate that makes the legacy read safe. A stale or hand-edited entry can
 * hold anything, and an unrepresentable mode in the store would reach the
 * derivation as a value it has no branch for.
 */
function isThemeMode(value: unknown): value is ThemeMode {
  return value === 'light' || value === 'dark' || value === 'system';
}

/**
 * The legacy mode held under `key`, or null when there is nothing usable to
 * carry across. The entry is read and left in place, never deleted: the old
 * store may still be mounted during the transition, and a stale key costs
 * nothing while a destructive read cannot be undone.
 *
 * Exported because there are two callers, and whoever changes the legacy shape
 * needs to know about both:
 *
 * 1. `createAppearanceStore` below, when `options.legacyModeKey` is given.
 * 2. `initAppearance` in `@voxire/ui`, which runs before this store is
 *    constructed. It has to resolve the mode itself to write the right palette
 *    on the first paint, so a user who already chose light does not get one
 *    frame of dark on the way in. It cannot read the store because the store
 *    does not exist yet.
 *
 * That second caller is the whole reason this is public rather than private.
 * Two parsers of the same persisted shape drift, and the drift shows up as a
 * flash of the wrong palette, which is the exact defect the pre-paint path
 * exists to prevent.
 */
export function readLegacyMode(key: string): ThemeMode | null {
  if (typeof window === 'undefined') return null;

  let raw: string | null = null;
  try {
    raw = window.localStorage.getItem(key);
  } catch {
    return null;
  }
  if (raw === null) return null;

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }
  if (typeof parsed !== 'object' || parsed === null) return null;

  const state = (parsed as { state?: unknown }).state;
  if (typeof state !== 'object' || state === null) return null;

  const theme = (state as { theme?: unknown }).theme;
  return isThemeMode(theme) ? theme : null;
}

/**
 * Create an appearance store persisted under `storageKey`.
 *
 * Precedence, highest first: the value persisted under `storageKey`, then the
 * mode migrated from `options.legacyModeKey` if one was given, then the
 * remaining `options` as defaults, then `APPEARANCE_DEFAULTS`. The first of
 * those is applied by persist's own hydration, which shallow-merges the stored
 * preferences over the initial state, so an already-migrated value is never
 * clobbered by the legacy read.
 */
export function createAppearanceStore(
  storageKey: string,
  options?: CreateAppearanceStoreOptions
): AppearanceStoreHook {
  const { legacyModeKey, ...defaults } = options ?? {};
  const base: AppearancePrefs = { ...APPEARANCE_DEFAULTS, ...defaults };
  const legacyMode = typeof legacyModeKey === 'string' ? readLegacyMode(legacyModeKey) : null;
  const initial: AppearancePrefs = legacyMode === null ? base : { ...base, mode: legacyMode };

  return create<AppearanceStore>()(
    persist(
      (set) => ({
        ...initial,
        setMode: (mode) => set({ mode }),
        setPalette: (palette) => set({ palette }),
        setAccent: (accent) => set({ accent }),
        setDensity: (density) => set({ density }),
        setTextSize: (textSize) => set({ textSize }),
        setCorners: (corners) => set({ corners }),
        setReduceMotion: (reduceMotion) => set({ reduceMotion }),
        reset: () => set({ ...base }),
      }),
      {
        name: storageKey,
        partialize: (state): AppearancePrefs => ({
          mode: state.mode,
          palette: state.palette,
          accent: state.accent,
          density: state.density,
          textSize: state.textSize,
          corners: state.corners,
          reduceMotion: state.reduceMotion,
        }),
      }
    )
  );
}
