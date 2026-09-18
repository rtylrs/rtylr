import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDebounce } from './useDebounce';

describe('useDebounce', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('hello', 300));
    expect(result.current).toBe('hello');
  });

  it('debounces value changes', () => {
    const { result, rerender } = renderHook(({ value, delay }) => useDebounce(value, delay), {
      initialProps: { value: 'a', delay: 300 },
    });

    expect(result.current).toBe('a');

    // Change value
    rerender({ value: 'ab', delay: 300 });
    expect(result.current).toBe('a'); // Still old value

    // Fast-forward 200ms (not enough)
    act(() => vi.advanceTimersByTime(200));
    expect(result.current).toBe('a');

    // Fast-forward remaining 100ms
    act(() => vi.advanceTimersByTime(100));
    expect(result.current).toBe('ab');
  });

  it('resets timer on rapid changes', () => {
    const { result, rerender } = renderHook(({ value }) => useDebounce(value, 300), {
      initialProps: { value: '' },
    });

    rerender({ value: 'a' });
    act(() => vi.advanceTimersByTime(100));
    rerender({ value: 'ab' });
    act(() => vi.advanceTimersByTime(100));
    rerender({ value: 'abc' });
    act(() => vi.advanceTimersByTime(100));

    // Only 100ms since last change - should still be empty
    expect(result.current).toBe('');

    // Wait full 300ms
    act(() => vi.advanceTimersByTime(200));
    expect(result.current).toBe('abc');
  });
});
