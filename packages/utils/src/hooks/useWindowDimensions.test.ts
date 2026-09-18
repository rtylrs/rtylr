import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWindowDimensions } from './useWindowDimensions';

describe('useWindowDimensions', () => {
  const originalInnerWidth = window.innerWidth;
  const originalInnerHeight = window.innerHeight;

  afterEach(() => {
    Object.defineProperty(window, 'innerWidth', { value: originalInnerWidth, writable: true });
    Object.defineProperty(window, 'innerHeight', { value: originalInnerHeight, writable: true });
  });

  it('returns current window dimensions', () => {
    Object.defineProperty(window, 'innerWidth', { value: 1024, writable: true });
    Object.defineProperty(window, 'innerHeight', { value: 768, writable: true });

    const { result } = renderHook(() => useWindowDimensions());

    expect(result.current.width).toBe(1024);
    expect(result.current.height).toBe(768);
  });

  it('returns isMobile false for desktop width', () => {
    Object.defineProperty(window, 'innerWidth', { value: 1024, writable: true });
    Object.defineProperty(window, 'innerHeight', { value: 768, writable: true });

    const { result } = renderHook(() => useWindowDimensions());

    expect(result.current.isMobile).toBe(false);
  });

  it('returns isMobile true for mobile width', () => {
    Object.defineProperty(window, 'innerWidth', { value: 375, writable: true });
    Object.defineProperty(window, 'innerHeight', { value: 667, writable: true });

    const { result } = renderHook(() => useWindowDimensions());

    expect(result.current.isMobile).toBe(true);
  });

  it('returns isMobile true at exactly 768px breakpoint', () => {
    Object.defineProperty(window, 'innerWidth', { value: 768, writable: true });
    Object.defineProperty(window, 'innerHeight', { value: 1024, writable: true });

    const { result } = renderHook(() => useWindowDimensions());

    expect(result.current.isMobile).toBe(true);
  });

  it('returns isMobile false at 769px', () => {
    Object.defineProperty(window, 'innerWidth', { value: 769, writable: true });
    Object.defineProperty(window, 'innerHeight', { value: 1024, writable: true });

    const { result } = renderHook(() => useWindowDimensions());

    expect(result.current.isMobile).toBe(false);
  });

  it('updates dimensions on window resize', async () => {
    Object.defineProperty(window, 'innerWidth', { value: 1024, writable: true });
    Object.defineProperty(window, 'innerHeight', { value: 768, writable: true });

    const { result } = renderHook(() => useWindowDimensions());

    act(() => {
      Object.defineProperty(window, 'innerWidth', { value: 500, writable: true });
      Object.defineProperty(window, 'innerHeight', { value: 800, writable: true });
      window.dispatchEvent(new Event('resize'));
    });

    expect(result.current.width).toBe(500);
    expect(result.current.height).toBe(800);
    expect(result.current.isMobile).toBe(true);
  });

  it('cleans up resize listener on unmount', () => {
    const removeSpy = vi.spyOn(window, 'removeEventListener');
    const { unmount } = renderHook(() => useWindowDimensions());
    unmount();

    expect(removeSpy).toHaveBeenCalledWith('resize', expect.any(Function));
    removeSpy.mockRestore();
  });

  it('returns numeric values', () => {
    const { result } = renderHook(() => useWindowDimensions());

    expect(typeof result.current.width).toBe('number');
    expect(typeof result.current.height).toBe('number');
    expect(typeof result.current.isMobile).toBe('boolean');
  });
});
