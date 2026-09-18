import React from 'react';
import { appearanceTokens, readLegacyMode, resolveMode } from '@voxire/utils';
import type {
  AppearanceContext,
  AppearancePrefs,
  Corners,
  Density,
  TextSize,
  ThemeMode,
} from '@voxire/utils';

/**
 * The one place appearance tokens reach the DOM.
 *
 * Two entry points, one writer. `initAppearance()` runs from main.tsx before
 * React mounts so the first paint is already the right palette; the provider
 * owns every write after that. Nothing else in any app may set these custom
 * properties on documentElement.
 *
 * That exclusivity is the point rather than a convention. An inline custom
 * property on documentElement outranks every stylesheet rule, so as long as a
 * layout writes accent tints inline, globals.css cannot theme them: a light-mode
 * value written in the stylesheet is silently lost the moment that layout
 * mounts. Flow hit exactly this and documented it in place, at
 * apps/flow/src/layouts/FlowLayout.tsx. Collapsing to a single writer removes
 * the conflict instead of adding another party to it.
 *
 * The colour derivation is not here. It lives in @voxire/utils as a pure
 * function so the contrast guarantees can be proven by a sweep instead of
 * asserted, and this file only decides where the result lands.
 *
 * The same division applies to the two rules this file used to own a copy of.
 * `resolveMode` and `readLegacyMode` are pure and live beside the store; what
 * belongs here is only the DOM read that feeds them. Two parsers of one
 * persisted shape drift, and the drift surfaces as a flash of the wrong palette,
 * which is the exact defect the pre-paint path exists to prevent.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr)
 */

const DARK_QUERY = '(prefers-color-scheme: dark)';
const REDUCED_MOTION_QUERY = '(prefers-reduced-motion: reduce)';

/** Attribute every `:root[data-theme='light']` rule in an app's CSS keys off. */
const THEME_ATTRIBUTE = 'data-theme';

/**
 * Attribute the reduced-motion CSS keys off. Written even when motion is on, so
 * a rule can select either state rather than relying on the absence of a value.
 */
const MOTION_ATTRIBUTE = 'data-motion';
const MOTION_REDUCE = 'reduce';
const MOTION_FULL = 'full';

const MODES: readonly ThemeMode[] = ['light', 'dark', 'system'];
const DENSITIES: readonly Density[] = ['compact', 'cozy', 'roomy'];
const TEXT_SIZES: readonly TextSize[] = ['s', 'm', 'l'];
const CORNERS: readonly Corners[] = ['sharp', 'soft', 'round'];

const HEX_PATTERN = /^#(?:[0-9a-f]{3}|[0-9a-f]{6}|[0-9a-f]{8})$/i;

function matches(query: string): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false;
  return window.matchMedia(query).matches;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/**
 * Narrows to one of `allowed`. The cast is needed because `includes` on a
 * readonly tuple of literals does not narrow an `unknown`, and it is safe
 * because the membership test is the narrowing.
 */
function isMember<T extends string>(value: unknown, allowed: readonly T[]): value is T {
  return typeof value === 'string' && (allowed as readonly string[]).includes(value);
}

function pick<T extends string>(value: unknown, allowed: readonly T[], fallback: T): T {
  return isMember(value, allowed) ? value : fallback;
}

function pickAccent(value: unknown, fallback: string | null): string | null {
  // An explicit null is a real choice: it means "follow the workspace colour".
  if (value === null) return null;
  if (typeof value === 'string' && HEX_PATTERN.test(value)) return value;
  return fallback;
}

function readStorage(key: string): unknown {
  // Everything here can fail for reasons that are not this app's problem:
  // Safari private mode throws on access, a half-written entry fails to parse,
  // and an older build may have stored a different shape entirely. A wrong
  // palette for one paint is a far better outcome than a blank screen.
  try {
    const raw = window.localStorage.getItem(key);
    if (!raw) return undefined;
    return JSON.parse(raw);
  } catch {
    return undefined;
  }
}

/**
 * Flattens whatever the preferences store persisted into one lookup.
 *
 * Read directly rather than through the store because this has to work before
 * the store is constructed. Three shapes are folded in, most specific last, so
 * this survives the store persisting the prefs at the root, under a zustand
 * `state` wrapper, or nested under `state.prefs`. Every field is validated
 * individually afterwards, so an unexpected shape degrades to the defaults one
 * field at a time instead of all at once.
 */
function flattenStored(parsed: unknown): Record<string, unknown> {
  const root = isRecord(parsed) ? parsed : {};
  const state = isRecord(root.state) ? root.state : {};
  const nested = isRecord(state.prefs) ? state.prefs : {};
  return { ...root, ...state, ...nested };
}

export interface AppearanceInitOptions {
  /** localStorage key the preferences store persists under. */
  storageKey: string;
  /**
   * The app's shipped defaults. Passed in rather than defined here so this
   * package carries no opinion about any one app's palette.
   */
  defaults: AppearancePrefs;
  /** Accent to derive from when `prefs.accent` is null. */
  workspaceAccent: string;
  /**
   * A legacy theme-only store's key, if the app has one. Only `mode` is taken
   * from it, and only when `storageKey` holds nothing, so a user who already
   * chose light does not get one paint of dark on the way to the migration the
   * store performs.
   *
   * `string | null` to match `CreateAppearanceStoreOptions.legacyModeKey`, so an
   * app threads one value to both without reshaping it.
   */
  legacyModeKey?: string | null;
}

function resolvePrefs(options: AppearanceInitOptions): AppearancePrefs {
  const { defaults, storageKey, legacyModeKey } = options;
  const raw = readStorage(storageKey);
  const stored = flattenStored(raw);

  let mode = pick(stored.mode, MODES, defaults.mode);
  if (raw === undefined && typeof legacyModeKey === 'string') {
    mode = readLegacyMode(legacyModeKey) ?? mode;
  }

  return {
    mode,
    palette: typeof stored.palette === 'string' ? stored.palette : defaults.palette,
    accent: pickAccent(stored.accent, defaults.accent),
    density: pick(stored.density, DENSITIES, defaults.density),
    textSize: pick(stored.textSize, TEXT_SIZES, defaults.textSize),
    corners: pick(stored.corners, CORNERS, defaults.corners),
    // No stored answer means the OS gets to answer. A stored `false` still
    // wins, because an explicit choice outranks an inferred one.
    reduceMotion:
      typeof stored.reduceMotion === 'boolean'
        ? stored.reduceMotion
        : defaults.reduceMotion || matches(REDUCED_MOTION_QUERY),
  };
}

/**
 * Writes the derived tokens and the two attributes. Returns the keys it wrote so
 * a caller that writes repeatedly can clear the ones that go away.
 */
function write(prefs: AppearancePrefs, context: AppearanceContext): string[] {
  const root = document.documentElement;
  const derived = appearanceTokens(prefs, context);

  for (const [name, value] of Object.entries(derived)) {
    root.style.setProperty(name, value);
  }
  root.setAttribute(THEME_ATTRIBUTE, context.resolvedMode);
  root.setAttribute(MOTION_ATTRIBUTE, prefs.reduceMotion ? MOTION_REDUCE : MOTION_FULL);

  return Object.keys(derived);
}

/**
 * Applies the persisted appearance before React renders.
 *
 * Call this from main.tsx, above `createRoot`. It exists so there is no flash of
 * the default palette: by the time the first paint happens, `data-theme` and
 * every token are already on documentElement. Nothing about it is allowed to
 * throw, since a failure here would take the app down before it has rendered
 * anything at all.
 */
export function initAppearance(options: AppearanceInitOptions): void {
  try {
    const prefs = resolvePrefs(options);
    write(prefs, {
      resolvedMode: resolveMode(prefs.mode, matches(DARK_QUERY)),
      workspaceAccent: options.workspaceAccent,
    });
  } catch {
    // Deliberately silent. The provider applies the same tokens on mount, so the
    // cost of a failure here is one paint of the stylesheet default, and the
    // alternative is a white screen before any error boundary exists.
  }
}

export interface AppearanceProviderProps {
  /**
   * The current preferences. Read from the app's own store and passed in, so
   * this package depends on no particular store instance.
   */
  prefs: AppearancePrefs;
  /** Accent to derive from when `prefs.accent` is null. */
  workspaceAccent: string;
  children?: React.ReactNode;
}

export function AppearanceProvider({
  prefs,
  workspaceAccent,
  children,
}: AppearanceProviderProps): React.JSX.Element {
  const { mode, palette, accent, density, textSize, corners, reduceMotion } = prefs;
  const isSystem = mode === 'system';

  const [systemDark, setSystemDark] = React.useState<boolean>(
    () => isSystem && matches(DARK_QUERY)
  );

  // Subscribed only while the mode is `system`. Gating the subscription rather
  // than the handler is what keeps an OS scheme change from re-writing tokens
  // for a user who picked light or dark explicitly.
  React.useEffect(() => {
    if (!isSystem) return;
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return;

    const query = window.matchMedia(DARK_QUERY);
    setSystemDark(query.matches);

    const onChange = (event: MediaQueryListEvent): void => setSystemDark(event.matches);
    query.addEventListener('change', onChange);
    return () => query.removeEventListener('change', onChange);
  }, [isSystem]);

  // `systemDark` is only ever true while subscribed, and resolveMode ignores it
  // for an explicit mode, so the gate above governs the subscription alone.
  const resolvedMode = resolveMode(mode, systemDark);

  // Rebuilt from the individual fields so a caller that hands us a fresh object
  // on every render does not re-run the write on every render.
  const current = React.useMemo<AppearancePrefs>(
    () => ({ mode, palette, accent, density, textSize, corners, reduceMotion }),
    [mode, palette, accent, density, textSize, corners, reduceMotion]
  );

  const writtenKeys = React.useRef<string[]>([]);

  React.useEffect(() => {
    const next = write(current, { resolvedMode, workspaceAccent });

    // Clear anything the previous derivation set and this one does not. Without
    // this, a token that drops out of the output keeps its stale inline value
    // and quietly outranks the stylesheet forever.
    const root = document.documentElement;
    for (const name of writtenKeys.current) {
      if (!next.includes(name)) root.style.removeProperty(name);
    }
    writtenKeys.current = next;
  }, [current, resolvedMode, workspaceAccent]);

  return <>{children}</>;
}

export default AppearanceProvider;
