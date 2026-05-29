import { describe, it, expect, beforeEach } from 'vitest';

import {
  getItem,
  setItem,
  removeItem,
  hasItem,
  flowKey,
  FLOW_STORAGE_PREFIX,
} from './storageUtils';

beforeEach(() => {
  localStorage.clear();
});

describe('setItem / getItem', () => {
  it('stores and retrieves a value', () => {
    setItem('testKey', { value: 42 });
    expect(getItem<{ value: number }>('testKey')).toEqual({ value: 42 });
  });

  it('returns null for missing key', () => {
    expect(getItem('missingKey')).toBeNull();
  });

  it('stores a string', () => {
    setItem('strKey', 'hello');
    expect(getItem<string>('strKey')).toBe('hello');
  });

  it('stores a number', () => {
    setItem('numKey', 99);
    expect(getItem<number>('numKey')).toBe(99);
  });

  it('stores an array', () => {
    setItem('arrKey', [1, 2, 3]);
    expect(getItem<number[]>('arrKey')).toEqual([1, 2, 3]);
  });
});

describe('removeItem', () => {
  it('removes a stored key', () => {
    setItem('toRemove', true);
    removeItem('toRemove');
    expect(getItem('toRemove')).toBeNull();
  });

  it('does not throw when key does not exist', () => {
    expect(() => removeItem('nonExistent')).not.toThrow();
  });
});

describe('hasItem', () => {
  it('returns true when key exists', () => {
    setItem('existKey', 'value');
    expect(hasItem('existKey')).toBe(true);
  });

  it('returns false when key does not exist', () => {
    expect(hasItem('nope')).toBe(false);
  });
});

describe('flowKey', () => {
  it('prepends flow prefix', () => {
    expect(flowKey('sidebar')).toBe(`${FLOW_STORAGE_PREFIX}sidebar`);
  });
});
