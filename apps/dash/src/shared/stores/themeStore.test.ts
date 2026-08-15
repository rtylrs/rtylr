import { describe, it, expect, beforeEach } from 'vitest';
import { useThemeStore, applyTheme, type Theme } from './themeStore';

const { getState } = useThemeStore;

beforeEach(() => {
  useThemeStore.setState({ theme: 'light' });
  document.documentElement.removeAttribute('data-theme');
  document.body.style.background = '';
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
