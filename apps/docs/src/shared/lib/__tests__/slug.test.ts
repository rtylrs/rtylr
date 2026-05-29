import { describe, it, expect } from 'vitest';
import { toSlug, slugToTitle, isValidSlug, uniqueSlug } from '../slug';

describe('toSlug', () => {
  it('lowercases and hyphenates', () => {
    expect(toSlug('Hello World')).toBe('hello-world');
  });

  it('removes special characters', () => {
    expect(toSlug('What is OAuth?')).toBe('what-is-oauth');
  });

  it('trims leading and trailing hyphens', () => {
    expect(toSlug('  My Title  ')).toBe('my-title');
  });

  it('collapses multiple spaces', () => {
    expect(toSlug('a   b')).toBe('a-b');
  });
});

describe('slugToTitle', () => {
  it('capitalizes each word', () => {
    expect(slugToTitle('getting-started')).toBe('Getting Started');
  });

  it('handles single word', () => {
    expect(slugToTitle('introduction')).toBe('Introduction');
  });
});

describe('isValidSlug', () => {
  it('returns true for valid slugs', () => {
    expect(isValidSlug('my-page')).toBe(true);
    expect(isValidSlug('api-v2-guide')).toBe(true);
  });

  it('returns false for uppercase', () => {
    expect(isValidSlug('My-Page')).toBe(false);
  });

  it('returns false for leading hyphen', () => {
    expect(isValidSlug('-my-page')).toBe(false);
  });

  it('returns false for trailing hyphen', () => {
    expect(isValidSlug('my-page-')).toBe(false);
  });
});

describe('uniqueSlug', () => {
  it('returns base slug if not in existing list', () => {
    expect(uniqueSlug('My Page', [])).toBe('my-page');
  });

  it('adds numeric suffix for duplicates', () => {
    expect(uniqueSlug('My Page', ['my-page'])).toBe('my-page-2');
  });

  it('increments suffix until unique', () => {
    expect(uniqueSlug('My Page', ['my-page', 'my-page-2'])).toBe('my-page-3');
  });
});
