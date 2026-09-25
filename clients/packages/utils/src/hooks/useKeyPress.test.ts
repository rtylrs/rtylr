import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useKeyPress, useKeyPressCallback } from './useKeyPress';

function fireKeydown(key: string) {
  window.dispatchEvent(new KeyboardEvent('keydown', { key }));
}
function fireKeyup(key: string) {
  window.dispatchEvent(new KeyboardEvent('keyup', { key }));
}

describe('useKeyPress', () => {
  it('returns false initially', () => {
    const { result } = renderHook(() => useKeyPress('Enter'));
    expect(result.current).toBe(false);
  });

  it('returns true when target key is pressed', () => {
    const { result } = renderHook(() => useKeyPress('Enter'));

    act(() => fireKeydown('Enter'));
    expect(result.current).toBe(true);
  });

  it('returns false when target key is released', () => {
    const { result } = renderHook(() => useKeyPress('Enter'));

    act(() => fireKeydown('Enter'));
    expect(result.current).toBe(true);

    act(() => fireKeyup('Enter'));
    expect(result.current).toBe(false);
  });

  it('does not respond to non-target keys', () => {
    const { result } = renderHook(() => useKeyPress('Enter'));

    act(() => fireKeydown('Escape'));
    expect(result.current).toBe(false);
  });

  it('handles Escape key', () => {
    const { result } = renderHook(() => useKeyPress('Escape'));

    act(() => fireKeydown('Escape'));
    expect(result.current).toBe(true);

    act(() => fireKeyup('Escape'));
    expect(result.current).toBe(false);
  });

  it('handles single character keys', () => {
    const { result } = renderHook(() => useKeyPress('a'));

    act(() => fireKeydown('a'));
    expect(result.current).toBe(true);

    act(() => fireKeyup('a'));
    expect(result.current).toBe(false);
  });

  it('cleans up event listeners on unmount', () => {
    const removeSpy = vi.spyOn(window, 'removeEventListener');
    const { unmount } = renderHook(() => useKeyPress('Enter'));
    unmount();

    expect(removeSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
    expect(removeSpy).toHaveBeenCalledWith('keyup', expect.any(Function));
    removeSpy.mockRestore();
  });

  it('does not respond after unmount', () => {
    const { result, unmount } = renderHook(() => useKeyPress('Enter'));
    unmount();

    act(() => fireKeydown('Enter'));
    expect(result.current).toBe(false);
  });
});

describe('useKeyPressCallback', () => {
  it('calls callback when target key is pressed', () => {
    const callback = vi.fn();
    renderHook(() => useKeyPressCallback('Enter', callback));

    act(() => fireKeydown('Enter'));
    expect(callback).toHaveBeenCalledTimes(1);
  });

  it('does not call callback for non-target keys', () => {
    const callback = vi.fn();
    renderHook(() => useKeyPressCallback('Enter', callback));

    act(() => fireKeydown('Escape'));
    expect(callback).not.toHaveBeenCalled();
  });

  it('calls callback multiple times on repeated presses', () => {
    const callback = vi.fn();
    renderHook(() => useKeyPressCallback('Enter', callback));

    act(() => {
      fireKeydown('Enter');
      fireKeydown('Enter');
      fireKeydown('Enter');
    });
    expect(callback).toHaveBeenCalledTimes(3);
  });

  it('stops calling after unmount', () => {
    const callback = vi.fn();
    const { unmount } = renderHook(() => useKeyPressCallback('Enter', callback));
    unmount();

    act(() => fireKeydown('Enter'));
    expect(callback).not.toHaveBeenCalled();
  });

  it('works with default empty dependencies', () => {
    const callback = vi.fn();
    renderHook(() => useKeyPressCallback('a', callback));

    act(() => fireKeydown('a'));
    expect(callback).toHaveBeenCalledTimes(1);
  });
});
