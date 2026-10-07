import { describe, it, expect } from 'vitest';
import { unwrapList } from './unwrapList';

// ── unwrapList ────────────────────────────────────────────────────────────────

describe('unwrapList', () => {
  // ── passthrough when already an array ────────────────────────────────────

  it('returns the array unchanged when input is already an array', () => {
    const input = [{ id: 1 }, { id: 2 }];
    expect(unwrapList(input, 'items')).toStrictEqual(input);
  });

  it('returns an empty array unchanged', () => {
    expect(unwrapList([], 'items')).toStrictEqual([]);
  });

  // ── single-key envelope ───────────────────────────────────────────────────

  it('unwraps a matching single-key envelope', () => {
    const payload = { customers: [{ id: 1 }, { id: 2 }] };
    expect(unwrapList(payload, 'customers')).toStrictEqual([{ id: 1 }, { id: 2 }]);
  });

  it('returns an empty array when the key exists but holds an empty array', () => {
    expect(unwrapList({ segments: [] }, 'segments')).toStrictEqual([]);
  });

  // ── multi-key fallback ────────────────────────────────────────────────────

  it('picks the first matching key when multiple candidate keys are supplied', () => {
    const payload = { campaigns: [{ id: 10 }], items: [{ id: 99 }] };
    // 'campaigns' is listed first - should win.
    expect(unwrapList(payload, 'campaigns', 'items')).toStrictEqual([{ id: 10 }]);
  });

  it('falls back to the second key when the first does not exist', () => {
    const payload = { items: [{ id: 99 }] };
    expect(unwrapList(payload, 'campaigns', 'items')).toStrictEqual([{ id: 99 }]);
  });

  // ── edge cases / missing keys ─────────────────────────────────────────────

  it('rejects a success response when no candidate key matches', () => {
    const payload = { other: [{ id: 1 }] };
    expect(() => unwrapList(payload, 'campaigns', 'items')).toThrow(
      'Expected a list response with "campaigns" or "items"'
    );
  });

  it('rejects null, missing, and primitive list responses', () => {
    for (const payload of [null, undefined, 'not-an-object', 42]) {
      expect(() => unwrapList(payload, 'customers')).toThrow(
        'Expected a list response with "customers"'
      );
    }
  });

  it('rejects a candidate key whose value is not an array', () => {
    expect(() => unwrapList({ customers: 'not-an-array' }, 'customers')).toThrow(
      'Expected a list response with "customers"'
    );
    expect(() => unwrapList({ customers: null }, 'customers')).toThrow(
      'Expected a list response with "customers"'
    );
  });

  // ── typed generic ─────────────────────────────────────────────────────────

  it('returns correctly typed items', () => {
    interface Item {
      id: number;
      name: string;
    }
    const payload = { items: [{ id: 1, name: 'Alpha' }] };
    const result = unwrapList<Item>(payload, 'items');
    expect(result[0].name).toBe('Alpha');
  });
});
