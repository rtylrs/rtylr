import { describe, it, expect, beforeEach } from 'vitest';
import { useFullscreenStore } from './fullscreenStore';

const { getState } = useFullscreenStore;

beforeEach(() => {
  useFullscreenStore.setState({ isFullscreen: true });
});

describe('fullscreenStore', () => {
  it('defaults to fullscreen true', () => {
    expect(getState().isFullscreen).toBe(true);
  });

  it('setFullscreen sets to false', () => {
    getState().setFullscreen(false);
    expect(getState().isFullscreen).toBe(false);
  });

  it('setFullscreen sets to true', () => {
    getState().setFullscreen(false);
    getState().setFullscreen(true);
    expect(getState().isFullscreen).toBe(true);
  });

  it('toggleFullscreen switches from true to false', () => {
    getState().toggleFullscreen();
    expect(getState().isFullscreen).toBe(false);
  });

  it('toggleFullscreen switches from false to true', () => {
    getState().setFullscreen(false);
    getState().toggleFullscreen();
    expect(getState().isFullscreen).toBe(true);
  });

  it('toggleFullscreen twice returns to original', () => {
    getState().toggleFullscreen();
    getState().toggleFullscreen();
    expect(getState().isFullscreen).toBe(true);
  });

  it('setFullscreen with same value is idempotent', () => {
    getState().setFullscreen(true);
    getState().setFullscreen(true);
    expect(getState().isFullscreen).toBe(true);
  });

  it('toggle after setFullscreen works correctly', () => {
    getState().setFullscreen(false);
    getState().toggleFullscreen();
    expect(getState().isFullscreen).toBe(true);
  });
});
