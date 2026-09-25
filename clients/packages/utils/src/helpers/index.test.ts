import { describe, it, expect, vi, beforeEach } from 'vitest';
import { helpers } from './index';

describe('helpers', () => {
  // ── isEmpty ──────────────────────────────────────────────────

  describe('isEmpty', () => {
    it('returns true for empty object', () => {
      expect(helpers.isEmpty({})).toBe(true);
    });

    it('returns true for empty array', () => {
      expect(helpers.isEmpty([])).toBe(true);
    });

    it('returns true for empty string', () => {
      expect(helpers.isEmpty('')).toBe(true);
    });

    it('returns true for null', () => {
      expect(helpers.isEmpty(null)).toBe(true);
    });

    it('returns true for undefined', () => {
      expect(helpers.isEmpty(undefined)).toBe(true);
    });

    it('returns false for non-empty object', () => {
      expect(helpers.isEmpty({ a: 1 })).toBe(false);
    });

    it('returns false for non-empty array', () => {
      expect(helpers.isEmpty([1, 2])).toBe(false);
    });

    it('returns false for non-empty string', () => {
      expect(helpers.isEmpty('hello')).toBe(false);
    });
  });

  // ── isEqual ──────────────────────────────────────────────────

  describe('isEqual', () => {
    it('returns true for equal primitives', () => {
      expect(helpers.isEqual(1, 1)).toBe(true);
      expect(helpers.isEqual('a', 'a')).toBe(true);
    });

    it('returns false for different primitives', () => {
      expect(helpers.isEqual(1, 2)).toBe(false);
    });

    it('returns true for deeply equal objects', () => {
      expect(helpers.isEqual({ a: { b: 1 } }, { a: { b: 1 } })).toBe(true);
    });

    it('returns false for different objects', () => {
      expect(helpers.isEqual({ a: 1 }, { a: 2 })).toBe(false);
    });

    it('returns true for equal arrays', () => {
      expect(helpers.isEqual([1, 2, 3], [1, 2, 3])).toBe(true);
    });

    it('returns false for different arrays', () => {
      expect(helpers.isEqual([1, 2], [1, 3])).toBe(false);
    });
  });

  // ── generateId ──────────────────────────────────────────────

  describe('generateId', () => {
    it('generates string of default length 8', () => {
      const id = helpers.generateId();
      expect(id).toHaveLength(8);
    });

    it('generates string of custom length', () => {
      const id = helpers.generateId(16);
      expect(id).toHaveLength(16);
    });

    it('generates alphanumeric characters only', () => {
      const id = helpers.generateId(100);
      expect(id).toMatch(/^[A-Za-z0-9]+$/);
    });

    it('generates unique ids', () => {
      const ids = new Set(Array.from({ length: 100 }, () => helpers.generateId()));
      expect(ids.size).toBe(100);
    });

    it('handles length 1', () => {
      const id = helpers.generateId(1);
      expect(id).toHaveLength(1);
    });
  });

  // ── deepClone ──────────────────────────────────────────────

  describe('deepClone', () => {
    it('clones simple object', () => {
      const obj = { a: 1, b: 'hello' };
      const cloned = helpers.deepClone(obj);
      expect(cloned).toEqual(obj);
      expect(cloned).not.toBe(obj);
    });

    it('clones nested object', () => {
      const obj = { a: { b: { c: 1 } } };
      const cloned = helpers.deepClone(obj);
      expect(cloned).toEqual(obj);
      cloned.a.b.c = 999;
      expect(obj.a.b.c).toBe(1);
    });

    it('clones arrays', () => {
      const arr = [1, 2, [3, 4]];
      const cloned = helpers.deepClone(arr);
      expect(cloned).toEqual(arr);
      expect(cloned).not.toBe(arr);
    });

    it('clones Date objects', () => {
      const date = new Date(2024, 0, 1);
      const cloned = helpers.deepClone(date);
      expect(cloned).toEqual(date);
      expect(cloned).not.toBe(date);
    });

    it('handles null', () => {
      expect(helpers.deepClone(null)).toBeNull();
    });

    it('handles primitives', () => {
      expect(helpers.deepClone(42)).toBe(42);
      expect(helpers.deepClone('hello')).toBe('hello');
      expect(helpers.deepClone(true)).toBe(true);
    });
  });

  // ── retry ──────────────────────────────────────────────────

  describe('retry', () => {
    it('returns result on first success', async () => {
      const fn = vi.fn().mockResolvedValue('ok');
      const result = await helpers.retry(fn);
      expect(result).toBe('ok');
      expect(fn).toHaveBeenCalledTimes(1);
    });

    it('retries on failure and succeeds', async () => {
      const fn = vi.fn().mockRejectedValueOnce(new Error('fail1')).mockResolvedValueOnce('ok');

      const result = await helpers.retry(fn, 3, 1);
      expect(result).toBe('ok');
      expect(fn).toHaveBeenCalledTimes(2);
    });

    it('throws after max attempts', async () => {
      const fn = vi.fn().mockRejectedValue(new Error('always fails'));

      await expect(helpers.retry(fn, 2, 1)).rejects.toThrow('always fails');
      expect(fn).toHaveBeenCalledTimes(2);
    });

    it('retries exactly maxAttempts times', async () => {
      const fn = vi.fn().mockRejectedValue(new Error('fail'));

      await expect(helpers.retry(fn, 3, 1)).rejects.toThrow();
      expect(fn).toHaveBeenCalledTimes(3);
    });
  });

  // ── sleep ──────────────────────────────────────────────────

  describe('sleep', () => {
    it('resolves after specified time', async () => {
      vi.useFakeTimers();
      const promise = helpers.sleep(100);
      vi.advanceTimersByTime(100);
      await expect(promise).resolves.toBeUndefined();
      vi.useRealTimers();
    });
  });

  // ── get ──────────────────────────────────────────────────

  describe('get', () => {
    it('gets nested property', () => {
      const obj = { a: { b: { c: 42 } } };
      expect(helpers.get(obj, 'a.b.c')).toBe(42);
    });

    it('returns defaultValue for missing path', () => {
      const obj = { a: 1 };
      expect(helpers.get(obj, 'b.c', 'default')).toBe('default');
    });

    it('returns undefined for missing path without default', () => {
      const obj = { a: 1 };
      expect(helpers.get(obj, 'b.c')).toBeUndefined();
    });

    it('handles null in path', () => {
      const obj = { a: null };
      expect(helpers.get(obj, 'a.b', 'fallback')).toBe('fallback');
    });

    it('gets top-level property', () => {
      const obj = { name: 'test' };
      expect(helpers.get(obj, 'name')).toBe('test');
    });

    it('returns defaultValue for null object', () => {
      expect(helpers.get(null, 'a.b', 'default')).toBe('default');
    });
  });

  // ── set ──────────────────────────────────────────────────

  describe('set', () => {
    it('sets nested property', () => {
      const obj: any = {};
      helpers.set(obj, 'a.b.c', 42);
      expect(obj.a.b.c).toBe(42);
    });

    it('sets top-level property', () => {
      const obj: any = {};
      helpers.set(obj, 'name', 'test');
      expect(obj.name).toBe('test');
    });

    it('overwrites existing property', () => {
      const obj: any = { a: { b: 1 } };
      helpers.set(obj, 'a.b', 2);
      expect(obj.a.b).toBe(2);
    });

    it('prevents prototype pollution in get', () => {
      const obj = {};
      expect(helpers.get(obj, '__proto__.polluted', 'default')).toBe('default');
      expect(helpers.get(obj, 'constructor.prototype.polluted', 'default')).toBe('default');
    });

    it('creates intermediate objects', () => {
      const obj: any = {};
      helpers.set(obj, 'deep.nested.value', true);
      expect(obj.deep.nested.value).toBe(true);
    });

    it('prevents prototype pollution in set', () => {
      const obj: any = {};
      helpers.set(obj, '__proto__.polluted', true);
      helpers.set(obj, 'constructor.prototype.polluted', true);
      expect(({} as any).polluted).toBeUndefined();
    });
  });

  // ── debounce ──────────────────────────────────────────────

  describe('debounce', () => {
    it('returns a function', () => {
      const fn = vi.fn();
      const debounced = helpers.debounce(fn, 100);
      expect(typeof debounced).toBe('function');
    });
  });

  // ── throttle ──────────────────────────────────────────────

  describe('throttle', () => {
    it('returns a function', () => {
      const fn = vi.fn();
      const throttled = helpers.throttle(fn, 100);
      expect(typeof throttled).toBe('function');
    });
  });
});
