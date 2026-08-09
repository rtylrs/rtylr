import { describe, it, expect } from 'vitest';
import { toBase64, fromBase64, toBase64Url, fromBase64Url, hashString } from '../crypto';

describe('toBase64 / fromBase64', () => {
  it('round-trips a string', () => {
    const original = 'hello world';
    expect(fromBase64(toBase64(original))).toBe(original);
  });

  it('encodes to non-empty string', () => {
    expect(toBase64('test').length).toBeGreaterThan(0);
  });
});

describe('toBase64Url / fromBase64Url', () => {
  it('round-trips a string', () => {
    const original = 'user:password123';
    expect(fromBase64Url(toBase64Url(original))).toBe(original);
  });

  it('contains no URL-unsafe characters', () => {
    const encoded = toBase64Url('some+data/here==');
    expect(encoded).not.toContain('+');
    expect(encoded).not.toContain('/');
    expect(encoded).not.toContain('=');
  });
});

describe('hashString', () => {
  it('returns a number', () => {
    expect(typeof hashString('hello')).toBe('number');
  });

  it('returns the same hash for the same input', () => {
    expect(hashString('test')).toBe(hashString('test'));
  });

  it('returns different hashes for different inputs', () => {
    expect(hashString('hello')).not.toBe(hashString('world'));
  });

  it('handles empty string', () => {
    expect(typeof hashString('')).toBe('number');
  });
});
