import { describe, it, expect, beforeEach } from 'vitest';
import { useFullscreenStore } from './fullscreenStore';

const { getState, setState } = useFullscreenStore;
const act = (fn: () => void) => {
  fn();
};

beforeEach(() => {
  setState({ isFullscreen: true });
});

describe('fullscreenStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts in fullscreen mode', () => {
    expect(getState().isFullscreen).toBe(true);
  });

  // ── setFullscreen ──────────────────────────────────────────────

  it('sets fullscreen to false', () => {
    act(() => getState().setFullscreen(false));
    expect(getState().isFullscreen).toBe(false);
  });

  it('sets fullscreen to true', () => {
    setState({ isFullscreen: false });
    act(() => getState().setFullscreen(true));
    expect(getState().isFullscreen).toBe(true);
  });

  it('sets fullscreen to same value (no-op)', () => {
    act(() => getState().setFullscreen(true));
    expect(getState().isFullscreen).toBe(true);
  });

  // ── toggleFullscreen ───────────────────────────────────────────

  it('toggles from fullscreen to windowed', () => {
    act(() => getState().toggleFullscreen());
    expect(getState().isFullscreen).toBe(false);
  });

  it('toggles from windowed to fullscreen', () => {
    setState({ isFullscreen: false });
    act(() => getState().toggleFullscreen());
    expect(getState().isFullscreen).toBe(true);
  });

  it('toggles back and forth', () => {
    act(() => getState().toggleFullscreen());
    expect(getState().isFullscreen).toBe(false);
    act(() => getState().toggleFullscreen());
    expect(getState().isFullscreen).toBe(true);
    act(() => getState().toggleFullscreen());
    expect(getState().isFullscreen).toBe(false);
  });

  // ── Combined usage ─────────────────────────────────────────────

  it('setFullscreen after toggle works correctly', () => {
    act(() => getState().toggleFullscreen());
    expect(getState().isFullscreen).toBe(false);
    act(() => getState().setFullscreen(true));
    expect(getState().isFullscreen).toBe(true);
  });

  it('toggle after setFullscreen(false) returns to true', () => {
    act(() => getState().setFullscreen(false));
    act(() => getState().toggleFullscreen());
    expect(getState().isFullscreen).toBe(true);
  });

  it('multiple rapid toggles produce correct state', () => {
    for (let i = 0; i < 10; i++) {
      act(() => getState().toggleFullscreen());
    }
    // 10 toggles from true -> false (even number of toggles)
    expect(getState().isFullscreen).toBe(true);
  });
});
