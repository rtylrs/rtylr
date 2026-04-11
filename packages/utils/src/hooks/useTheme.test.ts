import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createThemeStore, applyTheme } from './useTheme';

beforeEach(() => {
  window.localStorage.clear();
  document.documentElement.removeAttribute('data-theme');
  document.documentElement.classList.remove('dark');
});

describe('createThemeStore', () => {
  it('creates a store with default light mode', () => {
    const useStore = createThemeStore('test-theme-1');
    const state = useStore.getState();

    expect(state.mode).toBe('light');
    expect(state.resolved).toBe('light');
  });

  it('creates a store with custom default mode', () => {
    const useStore = createThemeStore('test-theme-2', 'dark');
    const state = useStore.getState();

    expect(state.mode).toBe('dark');
    expect(state.resolved).toBe('dark');
  });

  it('setMode updates mode and resolved for light', () => {
    const useStore = createThemeStore('test-theme-3');

    useStore.getState().setMode('light');

    expect(useStore.getState().mode).toBe('light');
    expect(useStore.getState().resolved).toBe('light');
  });

  it('setMode updates mode and resolved for dark', () => {
    const useStore = createThemeStore('test-theme-4');

    useStore.getState().setMode('dark');

    expect(useStore.getState().mode).toBe('dark');
    expect(useStore.getState().resolved).toBe('dark');
  });

  it('setMode resolves system theme (mocked as light)', () => {
    const useStore = createThemeStore('test-theme-5');

    useStore.getState().setMode('system');

    expect(useStore.getState().mode).toBe('system');
    // matchMedia is mocked to return matches: false, so system resolves to light
    expect(useStore.getState().resolved).toBe('light');
  });

  it('toggle switches from light to dark', () => {
    const useStore = createThemeStore('test-theme-6');

    useStore.getState().toggle();

    expect(useStore.getState().mode).toBe('dark');
    expect(useStore.getState().resolved).toBe('dark');
  });

  it('toggle switches from dark to light', () => {
    const useStore = createThemeStore('test-theme-7', 'dark');

    useStore.getState().toggle();

    expect(useStore.getState().mode).toBe('light');
    expect(useStore.getState().resolved).toBe('light');
  });

  it('toggle twice returns to original', () => {
    const useStore = createThemeStore('test-theme-8');

    useStore.getState().toggle();
    useStore.getState().toggle();

    expect(useStore.getState().mode).toBe('light');
    expect(useStore.getState().resolved).toBe('light');
  });

  it('creates independent stores with different keys', () => {
    const storeA = createThemeStore('theme-a');
    const storeB = createThemeStore('theme-b');

    storeA.getState().setMode('dark');

    expect(storeA.getState().mode).toBe('dark');
    expect(storeB.getState().mode).toBe('light');
  });
});

describe('applyTheme', () => {
  it('sets data-theme attribute to light', () => {
    applyTheme('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('sets data-theme attribute to dark', () => {
    applyTheme('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('adds dark class for dark theme', () => {
    applyTheme('dark');
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('removes dark class for light theme', () => {
    document.documentElement.classList.add('dark');
    applyTheme('light');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });
});
