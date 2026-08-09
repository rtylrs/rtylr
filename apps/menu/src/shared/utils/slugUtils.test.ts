import { describe, it, expect } from 'vitest';
import { toSlug, isValidSlug, extractSlugFromPath } from './slugUtils';

describe('toSlug', () => {
  it('lowercases and replaces spaces with hyphens', () => {
    expect(toSlug('Hello World')).toBe('hello-world');
  });

  it('removes special characters', () => {
    expect(toSlug("Joe's Diner!")).toBe('joes-diner');
  });

  it('collapses multiple spaces into a single hyphen', () => {
    expect(toSlug('a   b')).toBe('a-b');
  });

  it('trims leading and trailing hyphens', () => {
    expect(toSlug('  pizza  ')).toBe('pizza');
  });

  it('handles already-valid slugs unchanged', () => {
    expect(toSlug('my-restaurant')).toBe('my-restaurant');
  });

  it('handles an empty string', () => {
    expect(toSlug('')).toBe('');
  });
});

describe('isValidSlug', () => {
  it('returns true for a simple slug', () => {
    expect(isValidSlug('hello')).toBe(true);
  });

  it('returns true for a hyphenated slug', () => {
    expect(isValidSlug('my-restaurant')).toBe(true);
  });

  it('returns true for a slug with numbers', () => {
    expect(isValidSlug('resto-123')).toBe(true);
  });

  it('returns false for a slug starting with a hyphen', () => {
    expect(isValidSlug('-bad')).toBe(false);
  });

  it('returns false for a slug ending with a hyphen', () => {
    expect(isValidSlug('bad-')).toBe(false);
  });

  it('returns false for a slug with uppercase letters', () => {
    expect(isValidSlug('Hello')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isValidSlug('')).toBe(false);
  });
});

describe('extractSlugFromPath', () => {
  it('extracts slug from a /b/:slug path', () => {
    expect(extractSlugFromPath('/b/my-restaurant')).toBe('my-restaurant');
  });

  it('extracts slug from a /b/:slug/... path', () => {
    expect(extractSlugFromPath('/b/my-restaurant/menu')).toBe('my-restaurant');
  });

  it('extracts slug from a simple /:slug path', () => {
    expect(extractSlugFromPath('/my-place')).toBe('my-place');
  });

  it('returns null for the root path', () => {
    expect(extractSlugFromPath('/')).toBeNull();
  });
});
