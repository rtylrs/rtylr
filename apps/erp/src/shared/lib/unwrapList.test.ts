import { describe, it, expect } from 'vitest';
import { unwrapList } from './unwrapList';

// ---------------------------------------------------------------------------
// unwrapList - normalises ERP envelope responses to a plain array
// ---------------------------------------------------------------------------

describe('unwrapList', () => {
  it('returns the array as-is when passed a raw array', () => {
    const input = [{ id: 1 }, { id: 2 }];
    expect(unwrapList(input, 'items')).toEqual(input);
  });

  it('extracts an array from a single-key envelope', () => {
    const envelope = { products: [{ id: 10 }, { id: 11 }] };
    expect(unwrapList(envelope, 'products')).toEqual([{ id: 10 }, { id: 11 }]);
  });

  it('extracts the first matching key when multiple candidate keys are passed', () => {
    const envelope = { suppliers: [{ id: 1 }], items: [{ id: 2 }] };
    // 'suppliers' is checked first and matches.
    expect(unwrapList(envelope, 'suppliers', 'items')).toEqual([{ id: 1 }]);
  });

  it('falls through to the second key when the first is absent', () => {
    const envelope = { orders: [{ id: 7 }] };
    expect(unwrapList(envelope, 'products', 'orders')).toEqual([{ id: 7 }]);
  });

  it('returns an empty array when no key matches', () => {
    const envelope = { something: 'else' };
    expect(unwrapList(envelope, 'products', 'items')).toEqual([]);
  });

  it('returns an empty array for null input', () => {
    expect(unwrapList(null, 'products')).toEqual([]);
  });

  it('returns an empty array for undefined input', () => {
    expect(unwrapList(undefined, 'products')).toEqual([]);
  });

  it('returns an empty array for a non-object primitive', () => {
    expect(unwrapList(42, 'products')).toEqual([]);
    expect(unwrapList('string', 'products')).toEqual([]);
  });

  it('returns an empty array when the matching key holds a non-array value', () => {
    const envelope = { products: 'not-an-array' };
    expect(unwrapList(envelope, 'products')).toEqual([]);
  });

  it('works with an empty envelope array value', () => {
    const envelope = { products: [] };
    expect(unwrapList(envelope, 'products')).toEqual([]);
  });

  it('passes the generic type through for typed usage', () => {
    type Item = { id: number; name: string };
    const envelope = { items: [{ id: 1, name: 'Bolt' }] };
    const result: Item[] = unwrapList<Item>(envelope, 'items');
    expect(result[0].name).toBe('Bolt');
  });
});
