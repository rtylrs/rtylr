import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useThemeStore, initTheme, DEFAULT_THEME } from './themeStore';

const { getState, setState } = useThemeStore;

beforeEach(() => {
  setState({ theme: DEFAULT_THEME });
  document.documentElement.removeAttribute('data-theme');
  localStorage.clear();
});

describe('themeStore', () => {
  // Dark is the mode Flow is designed in: every screen in the design set is drawn
  // dark, and globals.css resolves dark from a bare `:root`.
  it('has dark as the default theme', () => {
    expect(DEFAULT_THEME).toBe('dark');
    expect(getState().theme).toBe('dark');
  });

  it('setTheme updates theme to light', () => {
    getState().setTheme('light');
    expect(getState().theme).toBe('light');
  });

  it('setTheme updates theme to system', () => {
    getState().setTheme('system');
    expect(getState().theme).toBe('system');
  });

  it('setTheme applies data-theme attribute for light', () => {
    getState().setTheme('light');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('setTheme applies data-theme attribute for dark', () => {
    getState().setTheme('dark');
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('setTheme with system uses matchMedia to resolve', () => {
    // matchMedia is mocked to return matches: false in setup.ts (light)
    getState().setTheme('system');
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('setTheme can be called multiple times', () => {
    getState().setTheme('light');
    getState().setTheme('dark');
    getState().setTheme('light');
    expect(getState().theme).toBe('light');
  });

  it('initTheme applies dark when no stored value', () => {
    localStorage.clear();
    initTheme();
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('initTheme applies stored theme from localStorage', () => {
    localStorage.setItem('flow-theme', JSON.stringify({ state: { theme: 'light' } }));
    initTheme();
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('initTheme handles malformed localStorage gracefully', () => {
    localStorage.setItem('flow-theme', 'not-json');
    initTheme();
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('still honours a stored light preference, so existing users keep their choice', () => {
    localStorage.setItem('flow-theme', JSON.stringify({ state: { theme: 'light' } }));
    initTheme();
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });
});
