import { describe, it, expect, beforeEach } from 'vitest';
import { getItem, setItem, removeItem, namespacedKey, getWithDefault } from './localStorageUtils';

beforeEach(() => {
  localStorage.clear();
});

describe('setItem / getItem', () => {
  it('stores and retrieves a string value', () => {
    setItem('key', 'hello');
    expect(getItem<string>('key')).toBe('hello');
  });

  it('stores and retrieves an object', () => {
    setItem('obj', { a: 1, b: 'two' });
    expect(getItem<{ a: number; b: string }>('obj')).toEqual({ a: 1, b: 'two' });
  });

  it('stores and retrieves a number', () => {
    setItem('num', 42);
    expect(getItem<number>('num')).toBe(42);
  });

  it('stores and retrieves a boolean', () => {
    setItem('flag', true);
    expect(getItem<boolean>('flag')).toBe(true);
  });
});

describe('getItem', () => {
  it('returns null for a missing key', () => {
    expect(getItem('missing')).toBeNull();
  });

  it('returns null for invalid JSON', () => {
    localStorage.setItem('bad', '{not json}');
    expect(getItem('bad')).toBeNull();
  });
});

describe('removeItem', () => {
  it('removes a key and returns true', () => {
    setItem('k', 'v');
    expect(removeItem('k')).toBe(true);
    expect(getItem('k')).toBeNull();
  });

  it('returns true even for a non-existent key', () => {
    expect(removeItem('missing')).toBe(true);
  });
});

describe('namespacedKey', () => {
  it('joins namespace and key with a colon', () => {
    expect(namespacedKey('dash', 'theme')).toBe('dash:theme');
  });

  it('handles empty namespace', () => {
    expect(namespacedKey('', 'key')).toBe(':key');
  });
});

describe('getWithDefault', () => {
  it('returns the stored value when present', () => {
    setItem('theme', 'dark');
    expect(getWithDefault('theme', 'light')).toBe('dark');
  });

  it('returns the default when the key is missing', () => {
    expect(getWithDefault('missing', 'light')).toBe('light');
  });

  it('returns the default for false stored values', () => {
    setItem('flag', false);
    expect(getWithDefault<boolean>('flag', true)).toBe(false);
  });
});
