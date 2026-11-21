import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import {
  readUserProjectPref,
  useUserProjectPref,
  userProjectStorageKey,
  writeUserProjectPref,
} from './userProjectStorage';

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('user project storage', () => {
  it('keys a preference by view, user and project', () => {
    expect(userProjectStorageKey('board:view', '1', '2')).toBe('flow:board:view:1:2');
  });

  it('keeps one person one project apart from everyone else', () => {
    writeUserProjectPref('v', '1', '2', { a: 1 });
    expect(readUserProjectPref('v', '1', '2')).toEqual({ a: 1 });
    expect(readUserProjectPref('v', '5', '2')).toBeUndefined();
    expect(readUserProjectPref('v', '1', '3')).toBeUndefined();
    expect(readUserProjectPref('w', '1', '2')).toBeUndefined();
  });

  it('reads and writes nothing before the user or project is known', () => {
    writeUserProjectPref('v', null, '2', 1);
    writeUserProjectPref('v', '1', '', 1);
    expect(localStorage.length).toBe(0);
    expect(readUserProjectPref('v', null, '2')).toBeUndefined();
  });

  it('reads a value that is not JSON as nothing stored', () => {
    localStorage.setItem('flow:v:1:2', '{nope');
    expect(readUserProjectPref('v', '1', '2')).toBeUndefined();
  });

  it('survives blocked storage', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    expect(readUserProjectPref('v', '1', '2')).toBeUndefined();
    expect(() => writeUserProjectPref('v', '1', '2', 1)).not.toThrow();
  });
});

describe('useUserProjectPref', () => {
  const load = (stored: unknown): number => (typeof stored === 'number' ? stored : 0);

  it('loads, updates and saves the preference', () => {
    writeUserProjectPref('v', '1', '2', 4);
    const { result } = renderHook(() => useUserProjectPref('v', '1', '2', load));
    expect(result.current[0]).toBe(4);

    act(() => {
      result.current[1]((n) => n + 1);
      result.current[1]((n) => n + 1);
    });
    expect(result.current[0]).toBe(6);
    expect(readUserProjectPref('v', '1', '2')).toBe(6);
  });

  it('writes nothing when a change keeps the value', () => {
    const { result } = renderHook(() => useUserProjectPref('v', '1', '2', load));
    act(() => result.current[1]((n) => n));
    expect(localStorage.length).toBe(0);
  });

  it('reads the new scope when the user or project changes', () => {
    writeUserProjectPref('v', '1', '3', 9);
    const { result, rerender } = renderHook(
      ({ userId, projectId }: { userId: string | null; projectId: string }) =>
        useUserProjectPref('v', userId, projectId, load),
      { initialProps: { userId: null as string | null, projectId: '3' } }
    );
    expect(result.current[0]).toBe(0);
    rerender({ userId: '1', projectId: '3' });
    expect(result.current[0]).toBe(9);
    rerender({ userId: '1', projectId: '4' });
    expect(result.current[0]).toBe(0);
  });
});
