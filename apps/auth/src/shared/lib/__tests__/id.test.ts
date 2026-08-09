import { describe, it, expect } from 'vitest';
import { generateId, isUuid, shortId } from '../id';

describe('generateId', () => {
  it('returns a string', () => {
    expect(typeof generateId()).toBe('string');
  });

  it('generates unique IDs', () => {
    const ids = new Set(Array.from({ length: 100 }, () => generateId()));
    expect(ids.size).toBe(100);
  });

  it('returns a value that looks like a UUID', () => {
    const id = generateId();
    expect(id).toMatch(/^[0-9a-f-]+$/i);
  });
});

describe('isUuid', () => {
  it('returns true for a valid UUID v4', () => {
    expect(isUuid('f47ac10b-58cc-4372-a567-0e02b2c3d479')).toBe(true);
    expect(isUuid('550e8400-e29b-11e4-a716-446655440000')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isUuid('')).toBe(false);
  });

  it('returns false for a non-UUID string', () => {
    expect(isUuid('hello-world')).toBe(false);
  });
});

describe('shortId', () => {
  it('returns a string of the requested length', () => {
    expect(shortId(8).length).toBe(8);
    expect(shortId(12).length).toBe(12);
  });

  it('defaults to 8 characters', () => {
    expect(shortId().length).toBe(8);
  });

  it('contains only alphanumeric characters', () => {
    expect(shortId(20)).toMatch(/^[a-z0-9]+$/);
  });

  it('generates unique IDs', () => {
    const ids = new Set(Array.from({ length: 50 }, () => shortId()));
    expect(ids.size).toBeGreaterThan(40);
  });
});
