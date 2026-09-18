import { describe, it, expect, beforeEach } from 'vitest';

import {
  APPEARANCE_DEFAULTS,
  LEGACY_THEME_STORAGE_KEY,
  createAppearanceStore,
  readLegacyMode,
  resolveMode,
} from './appearanceStore';
import { PALETTES } from './palettes';
import type { AppearancePrefs } from './types';

/**
 * Every test builds its own store under its own key. Two stores sharing a key
 * would share persisted state, and the legacy-migration cases in particular
 * depend on starting from an empty storage.
 */
let keyCounter = 0;
function nextKey(): string {
  keyCounter += 1;
  return `test-appearance-${keyCounter}`;
}

/** Seed the new-format entry persist itself would have written. */
function seedStored(key: string, prefs: Partial<AppearancePrefs>): void {
  window.localStorage.setItem(key, JSON.stringify({ state: prefs, version: 0 }));
}

/** Seed the shape Flow's hand-rolled store persisted under `flow-theme`. */
function seedLegacy(theme: unknown): void {
  window.localStorage.setItem(LEGACY_THEME_STORAGE_KEY, JSON.stringify({ state: { theme } }));
}

/** A store that has opted in to the legacy read, as Flow's call site does. */
function withLegacy(
  key: string,
  extra?: Partial<AppearancePrefs>
): ReturnType<typeof createAppearanceStore> {
  return createAppearanceStore(key, { legacyModeKey: LEGACY_THEME_STORAGE_KEY, ...extra });
}

beforeEach(() => {
  window.localStorage.clear();
  document.documentElement.removeAttribute('data-theme');
  document.documentElement.removeAttribute('style');
});

describe('APPEARANCE_DEFAULTS', () => {
  it('defaults mode to dark', () => {
    // Flow is designed dark: every screen in the design set is drawn dark and a
    // bare :root in globals.css resolves to the dark palette.
    expect(APPEARANCE_DEFAULTS.mode).toBe('dark');
  });

  it('defaults accent to null so the workspace colour is followed', () => {
    expect(APPEARANCE_DEFAULTS.accent).toBeNull();
  });

  it('names a palette that exists in the dark family', () => {
    const preset = PALETTES.find((p) => p.id === APPEARANCE_DEFAULTS.palette);
    expect(preset).toBeDefined();
    expect(preset?.mode).toBe('dark');
  });

  it('has the full documented shape', () => {
    expect(APPEARANCE_DEFAULTS).toEqual({
      mode: 'dark',
      palette: 'midnight',
      accent: null,
      density: 'cozy',
      textSize: 'm',
      corners: 'soft',
      reduceMotion: false,
    });
  });
});

describe('createAppearanceStore defaults', () => {
  it('starts from APPEARANCE_DEFAULTS when nothing is stored', () => {
    const useStore = createAppearanceStore(nextKey());
    const state = useStore.getState();

    expect(state.mode).toBe('dark');
    expect(state.palette).toBe('midnight');
    expect(state.accent).toBeNull();
    expect(state.density).toBe('cozy');
    expect(state.textSize).toBe('m');
    expect(state.corners).toBe('soft');
    expect(state.reduceMotion).toBe(false);
  });

  it('applies caller-supplied defaults over the shipped ones', () => {
    // This is how "reduce motion defaults on when the system asks for it" is
    // expressed without the store reading a media query itself.
    const useStore = createAppearanceStore(nextKey(), { reduceMotion: true, textSize: 'l' });
    const state = useStore.getState();

    expect(state.reduceMotion).toBe(true);
    expect(state.textSize).toBe('l');
    expect(state.mode).toBe('dark');
  });

  it('keeps stores under different keys independent', () => {
    const storeA = createAppearanceStore(nextKey());
    const storeB = createAppearanceStore(nextKey());

    storeA.getState().setMode('light');

    expect(storeA.getState().mode).toBe('light');
    expect(storeB.getState().mode).toBe('dark');
  });
});

describe('createAppearanceStore persistence', () => {
  it('lets a stored value win over the default', () => {
    const key = nextKey();
    seedStored(key, { mode: 'light', palette: 'linen', accent: '#8b8cf5' });

    const state = createAppearanceStore(key).getState();

    expect(state.mode).toBe('light');
    expect(state.palette).toBe('linen');
    expect(state.accent).toBe('#8b8cf5');
  });

  it('lets a stored value win over a caller-supplied default', () => {
    const key = nextKey();
    seedStored(key, { reduceMotion: false });

    const state = createAppearanceStore(key, { reduceMotion: true }).getState();

    expect(state.reduceMotion).toBe(false);
  });

  it('writes a setter result back to storage under the given key', () => {
    const key = nextKey();
    const useStore = createAppearanceStore(key);

    useStore.getState().setAccent('#f0805a');

    const raw = window.localStorage.getItem(key);
    expect(raw).not.toBeNull();
    expect(JSON.parse(raw ?? '{}')).toMatchObject({ state: { accent: '#f0805a' } });
  });
});

describe('createAppearanceStore legacy migration', () => {
  it('ignores the legacy entry entirely when no legacyModeKey is given', () => {
    // Opting in has to be explicit. A hardcoded read would hand dash and menu
    // Flow's mode on their first init when they adopt this factory.
    seedLegacy('light');

    expect(createAppearanceStore(nextKey()).getState().mode).toBe('dark');
  });

  it('ignores the legacy entry when legacyModeKey is null', () => {
    seedLegacy('light');

    expect(createAppearanceStore(nextKey(), { legacyModeKey: null }).getState().mode).toBe('dark');
  });

  it('reads only the key it was given', () => {
    seedLegacy('light');

    const state = createAppearanceStore(nextKey(), {
      legacyModeKey: 'some-other-app-theme',
    }).getState();

    expect(state.mode).toBe('dark');
  });

  it("migrates from a key other than Flow's own", () => {
    window.localStorage.setItem('dash-theme', JSON.stringify({ state: { theme: 'light' } }));

    const state = createAppearanceStore(nextKey(), { legacyModeKey: 'dash-theme' }).getState();

    expect(state.mode).toBe('light');
  });

  it('carries a legacy light choice across', () => {
    seedLegacy('light');

    expect(withLegacy(nextKey()).getState().mode).toBe('light');
  });

  it('carries a legacy system choice across', () => {
    seedLegacy('system');

    expect(withLegacy(nextKey()).getState().mode).toBe('system');
  });

  it('carries a legacy dark choice across', () => {
    // The caller default is light here on purpose: with the shipped dark default
    // this case would pass even if migration were dead.
    seedLegacy('dark');

    expect(withLegacy(nextKey(), { mode: 'light' }).getState().mode).toBe('dark');
  });

  it('migrates the mode only and leaves the other fields at their defaults', () => {
    seedLegacy('light');

    const state = withLegacy(nextKey()).getState();

    expect(state.mode).toBe('light');
    expect(state.palette).toBe(APPEARANCE_DEFAULTS.palette);
    expect(state.accent).toBeNull();
    expect(state.density).toBe(APPEARANCE_DEFAULTS.density);
  });

  it('does not clobber an already-migrated new-format value', () => {
    const key = nextKey();
    seedLegacy('light');
    seedStored(key, { mode: 'dark' });

    expect(withLegacy(key).getState().mode).toBe('dark');
  });

  it('ignores a legacy value that is not a theme mode', () => {
    // The gate that makes migration safe. Without it a stale or hand-edited
    // entry would put an unrepresentable mode into the store.
    seedLegacy('purple');

    expect(withLegacy(nextKey()).getState().mode).toBe('dark');
  });

  it('ignores a legacy entry whose theme is missing', () => {
    window.localStorage.setItem(LEGACY_THEME_STORAGE_KEY, JSON.stringify({ state: {} }));

    expect(withLegacy(nextKey()).getState().mode).toBe('dark');
  });

  it('ignores a legacy entry that is not valid JSON', () => {
    window.localStorage.setItem(LEGACY_THEME_STORAGE_KEY, 'not json');

    expect(withLegacy(nextKey()).getState().mode).toBe('dark');
  });

  it('lets a caller-supplied default lose to the legacy value', () => {
    seedLegacy('light');

    expect(withLegacy(nextKey(), { mode: 'dark' }).getState().mode).toBe('light');
  });
});

describe('createAppearanceStore setters', () => {
  it('sets each field independently', () => {
    const useStore = createAppearanceStore(nextKey());
    const s = useStore.getState();

    s.setMode('system');
    s.setPalette('mist');
    s.setAccent('#5ed6a0');
    s.setDensity('roomy');
    s.setTextSize('s');
    s.setCorners('sharp');
    s.setReduceMotion(true);

    expect(useStore.getState()).toMatchObject({
      mode: 'system',
      palette: 'mist',
      accent: '#5ed6a0',
      density: 'roomy',
      textSize: 's',
      corners: 'sharp',
      reduceMotion: true,
    });
  });

  it('setAccent(null) returns to following the workspace colour', () => {
    const useStore = createAppearanceStore(nextKey());

    useStore.getState().setAccent('#a3d84b');
    useStore.getState().setAccent(null);

    expect(useStore.getState().accent).toBeNull();
  });
});

describe('createAppearanceStore reset', () => {
  it('returns every field to the defaults', () => {
    const useStore = createAppearanceStore(nextKey());
    const s = useStore.getState();

    s.setMode('light');
    s.setPalette('slate');
    s.setAccent('#f56d8f');
    s.setDensity('compact');
    s.setTextSize('l');
    s.setCorners('round');
    s.setReduceMotion(true);

    useStore.getState().reset();

    expect(useStore.getState()).toMatchObject({ ...APPEARANCE_DEFAULTS });
  });

  it('returns to the caller-supplied defaults, not the shipped ones', () => {
    const useStore = createAppearanceStore(nextKey(), { reduceMotion: true, corners: 'round' });

    useStore.getState().setReduceMotion(false);
    useStore.getState().setCorners('sharp');
    useStore.getState().reset();

    expect(useStore.getState().reduceMotion).toBe(true);
    expect(useStore.getState().corners).toBe('round');
  });

  it('discards a migrated legacy mode, which is a choice rather than a default', () => {
    seedLegacy('light');
    const useStore = withLegacy(nextKey());
    expect(useStore.getState().mode).toBe('light');

    useStore.getState().reset();

    expect(useStore.getState().mode).toBe('dark');
  });

  it('keeps the setters callable after a reset', () => {
    const useStore = createAppearanceStore(nextKey());

    useStore.getState().reset();
    useStore.getState().setMode('light');

    expect(useStore.getState().mode).toBe('light');
  });
});

describe('createAppearanceStore is pure state', () => {
  it('writes nothing to the document', () => {
    // The provider is the single DOM writer. If this store ever sets an
    // attribute or a custom property there are two writers, and the inline one
    // silently outranks every stylesheet rule in globals.css.
    const useStore = createAppearanceStore(nextKey());
    const s = useStore.getState();

    s.setMode('light');
    s.setAccent('#64c5b9');
    s.setReduceMotion(true);
    s.reset();

    expect(document.documentElement.getAttribute('data-theme')).toBeNull();
    expect(document.documentElement.getAttribute('style')).toBeNull();
    expect(document.documentElement.classList.length).toBe(0);
  });
});

describe('readLegacyMode', () => {
  // Called directly by initAppearance in @voxire/ui, which runs before the store
  // exists. These four are the cases that path leans on, pinned here so the
  // contract lives with the implementation rather than only in its suite.

  it('returns null when the key holds nothing', () => {
    expect(readLegacyMode('no-such-key')).toBeNull();
  });

  it('returns null for malformed JSON', () => {
    window.localStorage.setItem('legacy-malformed', '{not json');

    expect(readLegacyMode('legacy-malformed')).toBeNull();
  });

  it('returns null for a valid JSON object with no theme field', () => {
    window.localStorage.setItem('legacy-no-theme', JSON.stringify({ state: { other: 1 } }));

    expect(readLegacyMode('legacy-no-theme')).toBeNull();
  });

  it('returns null for a theme value that is not a ThemeMode', () => {
    window.localStorage.setItem('legacy-bad-theme', JSON.stringify({ state: { theme: 'purple' } }));

    expect(readLegacyMode('legacy-bad-theme')).toBeNull();
  });

  it('returns null for JSON that is not an object', () => {
    window.localStorage.setItem('legacy-scalar', '"light"');

    expect(readLegacyMode('legacy-scalar')).toBeNull();
  });

  it('returns null when state is present but not an object', () => {
    window.localStorage.setItem('legacy-scalar-state', JSON.stringify({ state: 'light' }));

    expect(readLegacyMode('legacy-scalar-state')).toBeNull();
  });

  it('returns each of the three modes when the entry is well formed', () => {
    window.localStorage.setItem('legacy-ok', JSON.stringify({ state: { theme: 'light' } }));
    expect(readLegacyMode('legacy-ok')).toBe('light');

    window.localStorage.setItem('legacy-ok', JSON.stringify({ state: { theme: 'dark' } }));
    expect(readLegacyMode('legacy-ok')).toBe('dark');

    window.localStorage.setItem('legacy-ok', JSON.stringify({ state: { theme: 'system' } }));
    expect(readLegacyMode('legacy-ok')).toBe('system');
  });

  it('reads the key it is given and nothing else', () => {
    seedLegacy('light');

    expect(readLegacyMode('some-other-key')).toBeNull();
    expect(readLegacyMode(LEGACY_THEME_STORAGE_KEY)).toBe('light');
  });

  it('writes nothing to storage or to the document', () => {
    seedLegacy('light');
    const before = window.localStorage.length;

    readLegacyMode(LEGACY_THEME_STORAGE_KEY);

    expect(window.localStorage.length).toBe(before);
    expect(window.localStorage.getItem(LEGACY_THEME_STORAGE_KEY)).not.toBeNull();
    expect(document.documentElement.getAttribute('data-theme')).toBeNull();
  });
});

describe('resolveMode', () => {
  it('returns light for light regardless of the system preference', () => {
    expect(resolveMode('light', false)).toBe('light');
    expect(resolveMode('light', true)).toBe('light');
  });

  it('returns dark for dark regardless of the system preference', () => {
    expect(resolveMode('dark', false)).toBe('dark');
    expect(resolveMode('dark', true)).toBe('dark');
  });

  it('follows the supplied system preference for system', () => {
    expect(resolveMode('system', true)).toBe('dark');
    expect(resolveMode('system', false)).toBe('light');
  });

  it('is pure: the same inputs give the same answer and nothing is stored', () => {
    const before = window.localStorage.length;

    expect(resolveMode('system', true)).toBe(resolveMode('system', true));

    expect(window.localStorage.length).toBe(before);
    expect(document.documentElement.getAttribute('data-theme')).toBeNull();
  });
});
