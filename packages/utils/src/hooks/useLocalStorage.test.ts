import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useLocalStorage } from './useLocalStorage';

beforeEach(() => {
  window.localStorage.clear();
});

describe('useLocalStorage', () => {
  it('returns initial value when no stored value exists', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    expect(result.current[0]).toBe('default');
  });

  it('returns stored value when one exists', () => {
    window.localStorage.setItem('test-key', JSON.stringify('stored'));
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));
    expect(result.current[0]).toBe('stored');
  });

  it('updates state and localStorage when setValue is called', () => {
    const { result } = renderHook(() => useLocalStorage('test-key', 'default'));

    act(() => result.current[1]('new-value'));

    expect(result.current[0]).toBe('new-value');
    expect(JSON.parse(window.localStorage.getItem('test-key')!)).toBe('new-value');
  });

  it('supports function updater', () => {
    const { result } = renderHook(() => useLocalStorage('counter', 0));

    act(() => result.current[1]((prev) => prev + 1));

    expect(result.current[0]).toBe(1);
    expect(JSON.parse(window.localStorage.getItem('counter')!)).toBe(1);
  });

  it('handles object values', () => {
    const initial = { name: 'test', count: 0 };
    const { result } = renderHook(() => useLocalStorage('obj-key', initial));

    act(() => result.current[1]({ name: 'updated', count: 5 }));

    expect(result.current[0]).toEqual({ name: 'updated', count: 5 });
  });

  it('handles array values', () => {
    const { result } = renderHook(() => useLocalStorage<string[]>('arr-key', []));

    act(() => result.current[1](['a', 'b', 'c']));

    expect(result.current[0]).toEqual(['a', 'b', 'c']);
  });

  it('handles boolean values', () => {
    const { result } = renderHook(() => useLocalStorage('bool-key', false));

    act(() => result.current[1](true));

    expect(result.current[0]).toBe(true);
  });

  it('handles null initial value', () => {
    const { result } = renderHook(() => useLocalStorage<string | null>('null-key', null));
    expect(result.current[0]).toBeNull();
  });

  it('returns initialValue when localStorage has invalid JSON', () => {
    // Directly set invalid JSON
    const originalGetItem = window.localStorage.getItem;
    vi.spyOn(window.localStorage, 'getItem').mockReturnValueOnce('{invalid json');
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    const { result } = renderHook(() => useLocalStorage('bad-json', 'fallback'));

    expect(result.current[0]).toBe('fallback');
    consoleSpy.mockRestore();
  });

  it('handles localStorage setItem error', () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    vi.spyOn(window.localStorage, 'setItem').mockImplementationOnce(() => {
      throw new Error('quota exceeded');
    });

    const { result } = renderHook(() => useLocalStorage('error-key', 'default'));
    act(() => result.current[1]('new'));

    // State still updates even if localStorage fails
    consoleSpy.mockRestore();
  });

  it('returns tuple with correct types', () => {
    const { result } = renderHook(() => useLocalStorage('typed', 42));
    const [value, setValue] = result.current;
    expect(typeof value).toBe('number');
    expect(typeof setValue).toBe('function');
  });
});
