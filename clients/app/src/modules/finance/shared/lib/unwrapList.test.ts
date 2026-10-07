import { describe, it, expect } from 'vitest';
import { unwrapList } from './unwrapList';

describe('unwrapList', () => {
  it('returns the array as-is when given a plain array', () => {
    const input = [{ id: 1 }, { id: 2 }];
    expect(unwrapList(input, 'items')).toEqual(input);
  });

  it('unwraps the first matching key from an object envelope', () => {
    const input = { invoices: [{ id: 'INV-1' }, { id: 'INV-2' }] };
    expect(unwrapList(input, 'invoices')).toEqual(input.invoices);
  });

  it('returns the first matching key when multiple keys are provided', () => {
    const input = { bills: [{ id: 'BILL-1' }], invoices: [{ id: 'INV-1' }] };
    // "bills" is listed first → should win.
    const result = unwrapList(input, 'bills', 'invoices');
    expect(result).toEqual(input.bills);
  });

  it('falls through to the second key when the first key is absent', () => {
    const input = { invoices: [{ id: 'INV-1' }] };
    const result = unwrapList(input, 'bills', 'invoices');
    expect(result).toEqual(input.invoices);
  });

  it('rejects a success response when no key matches', () => {
    expect(() => unwrapList({ other: 'value' }, 'bills')).toThrow(
      'Expected a list response with "bills"'
    );
  });

  it('rejects null, missing, and primitive responses', () => {
    for (const response of [null, undefined, 42, 'hello']) {
      expect(() => unwrapList(response, 'items')).toThrow('Expected a list response with "items"');
    }
  });

  it('rejects missing keys and keys whose value is not an array', () => {
    expect(() => unwrapList({ invoices: [1, 2] })).toThrow(
      'Expected a list response with an array'
    );
    expect(() => unwrapList({ invoices: 'bad' }, 'invoices')).toThrow(
      'Expected a list response with "invoices"'
    );
  });

  it('passes through an empty array envelope correctly', () => {
    const result = unwrapList({ expenses: [] }, 'expenses');
    expect(result).toEqual([]);
  });

  it('preserves element types', () => {
    const items = ['a', 'b', 'c'];
    const result = unwrapList<string>({ list: items }, 'list');
    expect(result).toEqual(items);
    result.forEach((v) => expect(typeof v).toBe('string'));
  });

  it('ignores object keys whose value is not an array', () => {
    const input = { count: 5, items: [1, 2] };
    // "count" is not an array, so it should fall through to "items".
    const result = unwrapList(input, 'count', 'items');
    expect(result).toEqual([1, 2]);
  });
});
