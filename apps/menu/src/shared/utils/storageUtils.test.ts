import { describe, it, expect, beforeEach } from 'vitest';
import {
  localGet,
  localSet,
  localRemove,
  sessionGet,
  sessionSet,
  sessionRemove,
} from './storageUtils';

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
});

describe('localGet', () => {
  it('returns the stored value', () => {
    localStorage.setItem('key', JSON.stringify(42));
    expect(localGet('key', 0)).toBe(42);
  });

  it('returns the fallback when key is missing', () => {
    expect(localGet('missing', 'default')).toBe('default');
  });

  it('returns the fallback when JSON is invalid', () => {
    localStorage.setItem('bad', 'not-json{');
    expect(localGet('bad', null)).toBeNull();
  });
});

describe('localSet', () => {
  it('stores a value and returns true', () => {
    const ok = localSet('num', 99);
    expect(ok).toBe(true);
    expect(localStorage.getItem('num')).toBe('99');
  });
});

describe('localRemove', () => {
  it('removes the key from localStorage', () => {
    localStorage.setItem('k', 'v');
    localRemove('k');
    expect(localStorage.getItem('k')).toBeNull();
  });
});

describe('sessionGet', () => {
  it('returns the stored value', () => {
    sessionStorage.setItem('key', JSON.stringify({ x: 1 }));
    expect(sessionGet('key', {})).toEqual({ x: 1 });
  });

  it('returns the fallback when key is missing', () => {
    expect(sessionGet('nope', 'fb')).toBe('fb');
  });
});

describe('sessionSet', () => {
  it('stores a value and returns true', () => {
    const ok = sessionSet('s', [1, 2, 3]);
    expect(ok).toBe(true);
    expect(sessionStorage.getItem('s')).toBe('[1,2,3]');
  });
});

describe('sessionRemove', () => {
  it('removes the key from sessionStorage', () => {
    sessionStorage.setItem('s', 'v');
    sessionRemove('s');
    expect(sessionStorage.getItem('s')).toBeNull();
  });
});
