import React from 'react';
import { act, render } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { appearanceTokens, readLegacyMode, resolveMode } from '@voxire/utils';
import type { AppearancePrefs, AppearanceTokens, ResolvedMode, ThemeMode } from '@voxire/utils';
import { AppearanceProvider, initAppearance } from './AppearanceProvider';

/**
 * @voxire/utils is mocked on purpose.
 *
 * All three imports are pure functions owned by other pieces of this change, and
 * each has its own suite there. Mocking them keeps this a test of the writer:
 * which keys land on documentElement, which attributes are set, when the media
 * listener is attached and detached, and which values get handed to the pure
 * functions. The rules themselves are not retested here.
 *
 * `readLegacyMode` is a plain stub, so the legacy tests below assert wiring only
 * and hold no second copy of the parser. `resolveMode` needs a working double,
 * because data-theme is meaningless without one; the argument assertions beside
 * each data-theme test are what actually pin the wiring, so a drifting double
 * cannot make a passing test vacuous.
 *
 * An integration check against the real implementations is still owed.
 */
vi.mock('@voxire/utils', () => ({
  appearanceTokens: vi.fn(),
  resolveMode: vi.fn(),
  readLegacyMode: vi.fn(),
}));

const tokens = vi.mocked(appearanceTokens);
const resolve = vi.mocked(resolveMode);
const legacy = vi.mocked(readLegacyMode);

const DEFAULT_TOKENS: AppearanceTokens = {
  '--accent': '#64c5b9',
  '--accent-fg': '#1f7d72',
  '--flow-on-accent': '#0a0a0a',
};

const DARK_QUERY = '(prefers-color-scheme: dark)';
const MOTION_QUERY = '(prefers-reduced-motion: reduce)';

const STORAGE_KEY = 'test-appearance';

const DEFAULTS: AppearancePrefs = {
  mode: 'dark',
  palette: 'midnight',
  accent: null,
  density: 'cozy',
  textSize: 'm',
  corners: 'soft',
  reduceMotion: false,
};

interface FakeMediaQueryList {
  matches: boolean;
  media: string;
  addEventListener: ReturnType<typeof vi.fn>;
  removeEventListener: ReturnType<typeof vi.fn>;
  emit: (matches: boolean) => void;
}

/**
 * A matchMedia stub that can be driven from a test. The shared test setup
 * installs one that always reports `matches: false` and swallows listeners, so
 * every test here installs its own.
 */
function installMatchMedia(initial: Record<string, boolean>): Map<string, FakeMediaQueryList> {
  const lists = new Map<string, FakeMediaQueryList>();

  const make = (query: string): FakeMediaQueryList => {
    const listeners = new Set<(event: MediaQueryListEvent) => void>();
    const list: FakeMediaQueryList = {
      matches: initial[query] ?? false,
      media: query,
      addEventListener: vi.fn((_type: string, listener: (event: MediaQueryListEvent) => void) => {
        listeners.add(listener);
      }),
      removeEventListener: vi.fn(
        (_type: string, listener: (event: MediaQueryListEvent) => void) => {
          listeners.delete(listener);
        }
      ),
      emit: (matches: boolean) => {
        list.matches = matches;
        for (const listener of listeners) {
          listener({ matches } as MediaQueryListEvent);
        }
      },
    };
    return list;
  };

  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    configurable: true,
    value: vi.fn((query: string) => {
      const existing = lists.get(query);
      if (existing) return existing;
      const created = make(query);
      lists.set(query, created);
      return created;
    }),
  });

  return lists;
}

function mql(lists: Map<string, FakeMediaQueryList>, query: string): FakeMediaQueryList {
  const list = lists.get(query);
  if (!list) throw new Error(`matchMedia was never called with ${query}`);
  return list;
}

function root(): HTMLElement {
  return document.documentElement;
}

function lastCall(): Parameters<typeof appearanceTokens> {
  const calls = tokens.mock.calls;
  if (calls.length === 0) throw new Error('appearanceTokens was never called');
  return calls[calls.length - 1];
}

function lastPrefs(): AppearancePrefs {
  return lastCall()[0];
}

beforeEach(() => {
  tokens.mockReset();
  tokens.mockImplementation(() => ({ ...DEFAULT_TOKENS }));
  resolve.mockReset();
  resolve.mockImplementation(
    (mode: ThemeMode, prefersDark: boolean): ResolvedMode =>
      mode === 'system' ? (prefersDark ? 'dark' : 'light') : mode
  );
  legacy.mockReset();
  legacy.mockReturnValue(null);
  window.localStorage.clear();
  root().removeAttribute('data-theme');
  root().removeAttribute('data-motion');
  root().removeAttribute('style');
  installMatchMedia({});
});

afterEach(() => {
  root().removeAttribute('data-theme');
  root().removeAttribute('data-motion');
  root().removeAttribute('style');
});

describe('initAppearance', () => {
  it('applies the defaults and does not throw when localStorage is empty', () => {
    expect(() =>
      initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' })
    ).not.toThrow();

    expect(tokens).toHaveBeenCalledTimes(1);
    expect(lastPrefs()).toEqual(DEFAULTS);
    expect(tokens.mock.calls[0][1]).toEqual({
      resolvedMode: 'dark',
      workspaceAccent: '#8b5cf6',
    });
    expect(root().style.getPropertyValue('--accent')).toBe('#64c5b9');
  });

  it('falls back to the defaults on malformed JSON without throwing', () => {
    window.localStorage.setItem(STORAGE_KEY, '{"state":{"mode":');

    expect(() =>
      initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' })
    ).not.toThrow();

    expect(lastPrefs()).toEqual(DEFAULTS);
    expect(root().getAttribute('data-theme')).toBe('dark');
  });

  it('falls back to the defaults when the stored payload is not an object', () => {
    window.localStorage.setItem(STORAGE_KEY, '"light"');

    expect(() =>
      initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' })
    ).not.toThrow();
    expect(lastPrefs()).toEqual(DEFAULTS);
  });

  it('does not throw when localStorage itself is unavailable', () => {
    const throwing = {
      getItem: (): string => {
        throw new Error('SecurityError');
      },
    };
    const original = window.localStorage;
    Object.defineProperty(window, 'localStorage', { value: throwing, configurable: true });

    try {
      expect(() =>
        initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' })
      ).not.toThrow();
      expect(lastPrefs()).toEqual(DEFAULTS);
    } finally {
      Object.defineProperty(window, 'localStorage', { value: original, configurable: true });
    }
  });

  it('sets data-theme before anything renders', () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ state: { mode: 'light' } }));

    expect(document.body.childElementCount).toBe(0);
    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });

    // The attribute is on the element and no React tree has mounted, which is
    // the whole point of a pre-render call: the first paint is already correct.
    expect(root().getAttribute('data-theme')).toBe('light');
    expect(root().style.getPropertyValue('--accent')).toBe('#64c5b9');
    expect(document.body.childElementCount).toBe(0);
  });

  it('reads valid stored fields and ignores invalid ones field by field', () => {
    window.localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({
        state: {
          mode: 'light',
          palette: 'paper',
          accent: '#123456',
          density: 'nonsense',
          textSize: 'l',
          corners: 'round',
          reduceMotion: 'yes',
        },
      })
    );

    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });

    expect(lastPrefs()).toEqual({
      mode: 'light',
      palette: 'paper',
      accent: '#123456',
      density: 'cozy',
      textSize: 'l',
      corners: 'round',
      reduceMotion: false,
    });
  });

  it('honours an explicitly stored null accent', () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ state: { accent: null } }));

    initAppearance({
      storageKey: STORAGE_KEY,
      defaults: { ...DEFAULTS, accent: '#ff0000' },
      workspaceAccent: '#8b5cf6',
    });

    expect(lastPrefs().accent).toBeNull();
  });

  it('hands resolveMode its own read of the dark media query', () => {
    const lists = installMatchMedia({ [DARK_QUERY]: true });
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ state: { mode: 'system' } }));

    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });
    expect(resolve).toHaveBeenLastCalledWith('system', true);

    mql(lists, DARK_QUERY).matches = false;
    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });
    expect(resolve).toHaveBeenLastCalledWith('system', false);
  });

  it('resolves mode "system" against the dark media query', () => {
    const lists = installMatchMedia({ [DARK_QUERY]: true });
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ state: { mode: 'system' } }));

    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });

    expect(root().getAttribute('data-theme')).toBe('dark');
    mql(lists, DARK_QUERY).matches = false;

    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });
    expect(root().getAttribute('data-theme')).toBe('light');
  });

  it('defaults reduceMotion on when the system asks for it and nothing is stored', () => {
    installMatchMedia({ [MOTION_QUERY]: true });

    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });

    expect(lastPrefs().reduceMotion).toBe(true);
    expect(root().getAttribute('data-motion')).toBe('reduce');
  });

  it('lets a stored reduceMotion of false beat the system preference', () => {
    installMatchMedia({ [MOTION_QUERY]: true });
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ state: { reduceMotion: false } }));

    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });

    expect(lastPrefs().reduceMotion).toBe(false);
    expect(root().getAttribute('data-motion')).toBe('full');
  });

  it('takes the mode from readLegacyMode when the new key is absent', () => {
    legacy.mockReturnValue('light');

    initAppearance({
      storageKey: STORAGE_KEY,
      defaults: DEFAULTS,
      workspaceAccent: '#8b5cf6',
      legacyModeKey: 'flow-theme',
    });

    expect(legacy).toHaveBeenCalledWith('flow-theme');
    expect(lastPrefs().mode).toBe('light');
    expect(root().getAttribute('data-theme')).toBe('light');
  });

  it('keeps the default when there is no legacy value to carry across', () => {
    legacy.mockReturnValue(null);

    initAppearance({
      storageKey: STORAGE_KEY,
      defaults: DEFAULTS,
      workspaceAccent: '#8b5cf6',
      legacyModeKey: 'flow-theme',
    });

    expect(legacy).toHaveBeenCalledWith('flow-theme');
    expect(lastPrefs().mode).toBe('dark');
  });

  it('does not read the legacy key when the new key holds something', () => {
    legacy.mockReturnValue('light');
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ state: { mode: 'dark' } }));

    initAppearance({
      storageKey: STORAGE_KEY,
      defaults: DEFAULTS,
      workspaceAccent: '#8b5cf6',
      legacyModeKey: 'flow-theme',
    });

    expect(legacy).not.toHaveBeenCalled();
    expect(lastPrefs().mode).toBe('dark');
  });

  it('does not read the legacy key when legacyModeKey is null', () => {
    legacy.mockReturnValue('light');

    initAppearance({
      storageKey: STORAGE_KEY,
      defaults: DEFAULTS,
      workspaceAccent: '#8b5cf6',
      legacyModeKey: null,
    });

    expect(legacy).not.toHaveBeenCalled();
    expect(lastPrefs().mode).toBe('dark');
  });

  it('does not read the legacy key when no legacyModeKey is given', () => {
    initAppearance({ storageKey: STORAGE_KEY, defaults: DEFAULTS, workspaceAccent: '#8b5cf6' });

    expect(legacy).not.toHaveBeenCalled();
  });
});

describe('AppearanceProvider', () => {
  it('applies every token key the derivation returns', () => {
    tokens.mockReturnValueOnce({
      '--accent': '#111111',
      '--accent-muted': 'rgba(1,2,3,0.4)',
      '--flow-u': '4px',
      '--flow-r': '10px',
    });

    render(
      <AppearanceProvider prefs={DEFAULTS} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(root().style.getPropertyValue('--accent')).toBe('#111111');
    expect(root().style.getPropertyValue('--accent-muted')).toBe('rgba(1,2,3,0.4)');
    expect(root().style.getPropertyValue('--flow-u')).toBe('4px');
    expect(root().style.getPropertyValue('--flow-r')).toBe('10px');
  });

  it('renders its children', () => {
    const { getByText } = render(
      <AppearanceProvider prefs={DEFAULTS} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(getByText('child')).toBeInTheDocument();
  });

  it('removes a token that the derivation stops returning', () => {
    tokens.mockReturnValue({ '--accent': '#111111', '--going-away': '#222222' });

    const { rerender } = render(
      <AppearanceProvider prefs={DEFAULTS} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );
    expect(root().style.getPropertyValue('--going-away')).toBe('#222222');

    tokens.mockReturnValue({ '--accent': '#111111' });
    rerender(
      <AppearanceProvider prefs={DEFAULTS} workspaceAccent="#333333">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(root().style.getPropertyValue('--going-away')).toBe('');
    expect(root().style.getPropertyValue('--accent')).toBe('#111111');
  });

  it('sets data-theme to light for an explicit light mode', () => {
    installMatchMedia({ [DARK_QUERY]: true });

    render(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'light' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(root().getAttribute('data-theme')).toBe('light');
    expect(resolve).toHaveBeenLastCalledWith('light', false);
  });

  it('sets data-theme to dark for an explicit dark mode', () => {
    installMatchMedia({ [DARK_QUERY]: false });

    render(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'dark' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(root().getAttribute('data-theme')).toBe('dark');
  });

  it('resolves mode "system" to dark when the OS prefers dark', () => {
    installMatchMedia({ [DARK_QUERY]: true });

    render(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'system' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(root().getAttribute('data-theme')).toBe('dark');
    expect(resolve).toHaveBeenLastCalledWith('system', true);
  });

  it('resolves mode "system" to light when the OS prefers light', () => {
    installMatchMedia({ [DARK_QUERY]: false });

    render(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'system' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(root().getAttribute('data-theme')).toBe('light');
    expect(resolve).toHaveBeenLastCalledWith('system', false);
  });

  it('re-applies when the OS scheme changes and the mode is "system"', () => {
    const lists = installMatchMedia({ [DARK_QUERY]: false });

    render(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'system' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );
    expect(root().getAttribute('data-theme')).toBe('light');

    const before = tokens.mock.calls.length;
    act(() => mql(lists, DARK_QUERY).emit(true));

    expect(root().getAttribute('data-theme')).toBe('dark');
    expect(resolve).toHaveBeenLastCalledWith('system', true);
    expect(tokens.mock.calls.length).toBeGreaterThan(before);
    expect(lastPrefs().mode).toBe('system');
    expect(lastCall()[1]).toEqual({
      resolvedMode: 'dark',
      workspaceAccent: '#8b5cf6',
    });
  });

  it('does not re-apply on an OS scheme change when the mode is explicit', () => {
    const lists = installMatchMedia({ [DARK_QUERY]: false });

    render(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'light' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );
    expect(root().getAttribute('data-theme')).toBe('light');

    const before = tokens.mock.calls.length;
    const list = lists.get(DARK_QUERY);
    act(() => list?.emit(true));

    expect(root().getAttribute('data-theme')).toBe('light');
    expect(tokens.mock.calls.length).toBe(before);
  });

  it('removes the media query listener on unmount', () => {
    const lists = installMatchMedia({ [DARK_QUERY]: true });

    const { unmount } = render(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'system' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    const list = mql(lists, DARK_QUERY);
    expect(list.addEventListener).toHaveBeenCalledTimes(1);
    expect(list.removeEventListener).not.toHaveBeenCalled();

    unmount();

    expect(list.removeEventListener).toHaveBeenCalledTimes(1);
    expect(list.removeEventListener.mock.calls[0][0]).toBe('change');
    expect(list.removeEventListener.mock.calls[0][1]).toBe(list.addEventListener.mock.calls[0][1]);
  });

  it('removes the listener when the mode stops being "system"', () => {
    const lists = installMatchMedia({ [DARK_QUERY]: true });

    const { rerender } = render(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'system' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );
    const list = mql(lists, DARK_QUERY);
    expect(list.removeEventListener).not.toHaveBeenCalled();

    rerender(
      <AppearanceProvider prefs={{ ...DEFAULTS, mode: 'light' }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(list.removeEventListener).toHaveBeenCalledTimes(1);
  });

  it('sets data-motion from reduceMotion', () => {
    const { rerender } = render(
      <AppearanceProvider prefs={{ ...DEFAULTS, reduceMotion: true }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );
    expect(root().getAttribute('data-motion')).toBe('reduce');

    rerender(
      <AppearanceProvider prefs={{ ...DEFAULTS, reduceMotion: false }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );
    expect(root().getAttribute('data-motion')).toBe('full');
  });

  it('passes the workspace accent through as the derivation fallback', () => {
    render(
      <AppearanceProvider prefs={{ ...DEFAULTS, accent: null }} workspaceAccent="#8b5cf6">
        <span>child</span>
      </AppearanceProvider>
    );

    expect(lastCall()[0].accent).toBeNull();
    expect(lastCall()[1].workspaceAccent).toBe('#8b5cf6');
  });
});
