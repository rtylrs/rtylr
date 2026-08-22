import {
  LEGACY_THEME_STORAGE_KEY,
  createAppearanceStore,
  type AppearanceStoreHook,
} from '@voxire/utils';

/**
 * Flow's appearance preferences.
 *
 * The store itself, the token derivation and the DOM writer all live in the
 * shared packages. This file is the only place that knows they are being wired
 * up for Flow specifically, which is what keeps `@voxire/utils` free of a
 * hardcoded `flow-theme` string.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md
 */

/** Namespaced so a future dash or menu store cannot collide with it. */
export const FLOW_APPEARANCE_STORAGE_KEY = 'flow-appearance';

/**
 * Read once at module load rather than per render. This is a DOM read, which is
 * why it happens here and not inside the store: the store is pure by design, so
 * the system preference has to be passed in by whoever constructs it.
 *
 * A stored preference always wins over this, so a user who has explicitly
 * turned reduced motion off keeps it off even on a system that asks for it.
 */
function prefersReducedMotion(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false;
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * `legacyModeKey` is what carries a user's existing light or dark choice across
 * from the retired `themeStore`. Without it, everyone who had chosen light would
 * silently be handed the dark default on the first load after this ships.
 */
export const useAppearanceStore: AppearanceStoreHook = createAppearanceStore(
  FLOW_APPEARANCE_STORAGE_KEY,
  {
    legacyModeKey: LEGACY_THEME_STORAGE_KEY,
    reduceMotion: prefersReducedMotion(),
  }
);
