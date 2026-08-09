import { describe, it, expect, beforeEach } from 'vitest';
import { useThemeStore } from './themeStore';

const { getState, setState } = useThemeStore;

beforeEach(() => {
  setState({ theme: 'dark' });
});

describe('menu themeStore', () => {
  it('has dark as the default theme', () => {
    expect(getState().theme).toBe('dark');
  });

  it('setTheme switches to light', () => {
    getState().setTheme('light');
    expect(getState().theme).toBe('light');
  });

  it('setTheme switches to dark', () => {
    getState().setTheme('light');
    getState().setTheme('dark');
    expect(getState().theme).toBe('dark');
  });

  it('toggleTheme switches from dark to light', () => {
    getState().toggleTheme();
    expect(getState().theme).toBe('light');
  });

  it('toggleTheme switches from light to dark', () => {
    setState({ theme: 'light' });
    getState().toggleTheme();
    expect(getState().theme).toBe('dark');
  });

  it('toggleTheme twice returns to original theme', () => {
    getState().toggleTheme();
    getState().toggleTheme();
    expect(getState().theme).toBe('dark');
  });

  it('setTheme overrides after toggle', () => {
    getState().toggleTheme();
    expect(getState().theme).toBe('light');
    getState().setTheme('dark');
    expect(getState().theme).toBe('dark');
  });

  it('toggle after setTheme works correctly', () => {
    getState().setTheme('light');
    getState().toggleTheme();
    expect(getState().theme).toBe('dark');
  });
});
