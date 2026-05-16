import { describe, expect, it } from 'vitest';
import { safeJsonParse, safeJsonStringify } from './safeJson';

describe('safeJsonParse', () => {
  it('returns parsed value for valid JSON', () => {
    expect(safeJsonParse('{"a":1}', null)).toEqual({ a: 1 });
  });
  it('returns the fallback for invalid JSON', () => {
    expect(safeJsonParse('not json', 'fb')).toBe('fb');
  });
});

describe('safeJsonStringify', () => {
  it('stringifies normal values', () => {
    expect(safeJsonStringify({ a: 1 })).toBe('{"a":1}');
  });
  it('returns fallback for circular structures', () => {
    const a: Record<string, unknown> = {};
    a.self = a;
    expect(safeJsonStringify(a)).toBe('');
  });
});
