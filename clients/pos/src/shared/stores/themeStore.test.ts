import { describe, it, expect, beforeEach } from 'vitest';
import { useThemeStore } from './themeStore';

const { getState, setState } = useThemeStore;
const act = (fn: () => void) => {
  fn();
};

beforeEach(() => {
  setState({ theme: 'dark' });
});

describe('themeStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts with dark theme', () => {
    expect(getState().theme).toBe('dark');
  });

  // ── toggleTheme ────────────────────────────────────────────────

  it('toggles from dark to light', () => {
    act(() => getState().toggleTheme());
    expect(getState().theme).toBe('light');
  });

  it('toggles from light to dark', () => {
    setState({ theme: 'light' });
    act(() => getState().toggleTheme());
    expect(getState().theme).toBe('dark');
  });

  it('toggles back and forth', () => {
    act(() => getState().toggleTheme());
    expect(getState().theme).toBe('light');
    act(() => getState().toggleTheme());
    expect(getState().theme).toBe('dark');
    act(() => getState().toggleTheme());
    expect(getState().theme).toBe('light');
  });

  // ── setTheme ───────────────────────────────────────────────────

  it('sets theme to light', () => {
    act(() => getState().setTheme('light'));
    expect(getState().theme).toBe('light');
  });

  it('sets theme to dark', () => {
    setState({ theme: 'light' });
    act(() => getState().setTheme('dark'));
    expect(getState().theme).toBe('dark');
  });

  it('sets an arbitrary theme name', () => {
    act(() => getState().setTheme('solarized'));
    expect(getState().theme).toBe('dark');
  });

  // ── Edge cases ─────────────────────────────────────────────────

  it('setTheme to same value is a no-op', () => {
    act(() => getState().setTheme('dark'));
    expect(getState().theme).toBe('dark');
  });

  it('toggle after setTheme works correctly', () => {
    act(() => getState().setTheme('light'));
    act(() => getState().toggleTheme());
    expect(getState().theme).toBe('dark');
  });

  it('toggle with custom dark theme goes to light', () => {
    act(() => getState().setTheme('aurora'));
    act(() => getState().toggleTheme());
    expect(getState().theme).toBe('light');
  });
});
