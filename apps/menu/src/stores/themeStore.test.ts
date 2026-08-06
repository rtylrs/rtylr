import { describe, it, expect, beforeEach } from 'vitest';
import { useThemeStore } from './themeStore';

const { getState, setState } = useThemeStore;

beforeEach(() => {
  setState({ theme: 'light' });
});

describe('menu themeStore', () => {
  it('has light as the default theme', () => {
    expect(getState().theme).toBe('light');
  });

  it('setTheme switches to dark', () => {
    getState().setTheme('dark');
    expect(getState().theme).toBe('dark');
  });

  it('setTheme switches to light', () => {
    getState().setTheme('dark');
    getState().setTheme('light');
    expect(getState().theme).toBe('light');
  });

  it('toggleTheme switches from light to dark', () => {
    getState().toggleTheme();
    expect(getState().theme).toBe('dark');
  });

  it('toggleTheme switches from dark to light', () => {
    setState({ theme: 'dark' });
    getState().toggleTheme();
    expect(getState().theme).toBe('light');
  });

  it('toggleTheme twice returns to original theme', () => {
    getState().toggleTheme();
    getState().toggleTheme();
    expect(getState().theme).toBe('light');
  });

  it('setTheme overrides after toggle', () => {
    getState().toggleTheme();
    expect(getState().theme).toBe('dark');
    getState().setTheme('light');
    expect(getState().theme).toBe('light');
  });

  it('toggle after setTheme works correctly', () => {
    getState().setTheme('dark');
    getState().toggleTheme();
    expect(getState().theme).toBe('light');
  });
});
