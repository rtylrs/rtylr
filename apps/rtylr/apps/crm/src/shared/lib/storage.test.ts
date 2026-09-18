import { afterEach, describe, expect, it, vi } from 'vitest';
import { readLocalStorage, writeLocalStorage } from './storage';

afterEach(() => {
  vi.restoreAllMocks();
  window.localStorage.clear();
});

describe('readLocalStorage', () => {
  it('returns a stored value', () => {
    window.localStorage.setItem('crm-test', 'value');

    expect(readLocalStorage('crm-test')).toBe('value');
  });

  it('returns null when storage access is blocked', () => {
    vi.spyOn(window.localStorage, 'getItem').mockImplementation(() => {
      throw new Error('storage blocked');
    });

    expect(readLocalStorage('crm-test')).toBeNull();
  });
});

describe('writeLocalStorage', () => {
  it('persists a value', () => {
    writeLocalStorage('crm-test', 'value');

    expect(window.localStorage.getItem('crm-test')).toBe('value');
  });

  it('does not throw when storage access is blocked', () => {
    vi.spyOn(window.localStorage, 'setItem').mockImplementation(() => {
      throw new Error('storage blocked');
    });

    expect(() => writeLocalStorage('crm-test', 'value')).not.toThrow();
  });
});
