import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useThemeStore, applyTheme, isSystemListenerActive, type Theme } from './themeStore';

const { getState } = useThemeStore;

const listeners = vi.hoisted(() => ({
  added: [] as (() => void)[],
  removed: [] as (() => void)[],
}));

function mockMatchMedia() {
  listeners.added = [];
  listeners.removed = [];
  Object.defineProperty(window, 'matchMedia', {
    configurable: true,
    writable: true,
    value: vi.fn().mockReturnValue({
      matches: false,
      addEventListener: vi.fn((_event: string, handler: () => void) => {
        listeners.added.push(handler);
      }),
      removeEventListener: vi.fn((_event: string, handler: () => void) => {
        listeners.removed.push(handler);
      }),
    }),
  });
}

beforeEach(() => {
  useThemeStore.setState({ theme: 'light' });
  document.documentElement.removeAttribute('data-theme');
  document.body.style.background = '';
  getState().setTheme('light');
});

describe('themeStore (dash)', () => {
  it('has correct initial state', () => {
    expect(getState().theme).toBe('light');
  });

  it('setTheme updates to dark', () => {
    getState().setTheme('dark');
    expect(getState().theme).toBe('dark');
  });

  it('setTheme updates to light', () => {
    getState().setTheme('dark');
    getState().setTheme('light');
    expect(getState().theme).toBe('light');
  });

  it('setTheme updates to system', () => {
    getState().setTheme('system');
    expect(getState().theme).toBe('system');
  });

  it('setTheme applies data-theme attribute', () => {
    getState().setTheme('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('setTheme clears inline body background for dark', () => {
    getState().setTheme('dark');
    expect(document.body.style.background).toBe('');
  });

  it('setTheme clears inline body background for light', () => {
    getState().setTheme('light');
    expect(document.body.style.background).toBe('');
  });
});

describe('applyTheme', () => {
  it('applies dark theme', () => {
    applyTheme('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    expect(document.body.style.background).toBe('');
  });

  it('applies light theme', () => {
    applyTheme('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
    expect(document.body.style.background).toBe('');
  });

  it('applies system theme (resolves to light when matchMedia returns false)', () => {
    applyTheme('system');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });
});

describe('system theme listener lifecycle', () => {
  beforeEach(() => {
    mockMatchMedia();
  });

  it('attaches the OS listener when switching to system', () => {
    getState().setTheme('system');

    expect(isSystemListenerActive()).toBe(true);
    expect(listeners.added).toHaveLength(1);
  });

  it('detaches the OS listener when leaving system', () => {
    getState().setTheme('system');
    const [handler] = listeners.added;

    getState().setTheme('light');

    expect(isSystemListenerActive()).toBe(false);
    expect(listeners.removed).toEqual([handler]);
  });

  it('replaces rather than stacks the listener on repeat system sets', () => {
    getState().setTheme('system');
    getState().setTheme('system');

    expect(listeners.added).toHaveLength(2);
    expect(listeners.removed).toHaveLength(1);
    expect(isSystemListenerActive()).toBe(true);
  });
});
