import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  hasLocalStorage,
  storageGet,
  storageSet,
  storageRemove,
  storageGetJson,
  storageSetJson,
} from '../storage';

const mockStorage: Record<string, string> = {};

beforeEach(() => {
  Object.keys(mockStorage).forEach((k) => delete mockStorage[k]);
  vi.stubGlobal('window', {
    localStorage: {
      getItem: (key: string) => mockStorage[key] ?? null,
      setItem: (key: string, value: string) => {
        mockStorage[key] = value;
      },
      removeItem: (key: string) => {
        delete mockStorage[key];
      },
    },
  });
});

describe('hasLocalStorage', () => {
  it('returns true when localStorage is present', () => {
    expect(hasLocalStorage()).toBe(true);
  });
});

describe('storageGet / storageSet', () => {
  it('sets and gets a string value', () => {
    storageSet('foo', 'bar');
    expect(storageGet('foo')).toBe('bar');
  });

  it('returns null for missing keys', () => {
    expect(storageGet('missing')).toBeNull();
  });
});

describe('storageRemove', () => {
  it('removes an existing key', () => {
    storageSet('key', 'val');
    storageRemove('key');
    expect(storageGet('key')).toBeNull();
  });
});

describe('storageGetJson', () => {
  it('parses a stored JSON value', () => {
    storageSet('obj', JSON.stringify({ x: 1 }));
    expect(storageGetJson('obj', null)).toEqual({ x: 1 });
  });

  it('returns the fallback for missing keys', () => {
    expect(storageGetJson('no-key', 42)).toBe(42);
  });

  it('returns the fallback for invalid JSON', () => {
    storageSet('bad', '{not json}');
    expect(storageGetJson('bad', 'default')).toBe('default');
  });
});

describe('storageSetJson', () => {
  it('serialises an object to localStorage', () => {
    storageSetJson('data', { a: 1 });
    expect(storageGet('data')).toBe(JSON.stringify({ a: 1 }));
  });
});
