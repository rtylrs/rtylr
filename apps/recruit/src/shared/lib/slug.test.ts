import { describe, it, expect } from 'vitest';
import { toSlug, isValidSlug, slugToTitle, careerUrl, applyUrl } from './slug';

describe('toSlug', () => {
  it('lowercases and replaces spaces with hyphens', () => {
    expect(toSlug('Senior Engineer')).toBe('senior-engineer');
  });

  it('strips special characters', () => {
    expect(toSlug('C++ Developer!')).toBe('c-developer');
  });

  it('trims leading and trailing hyphens', () => {
    expect(toSlug('  hello world  ')).toBe('hello-world');
  });

  it('collapses multiple separators', () => {
    expect(toSlug('foo   bar--baz')).toBe('foo-bar-baz');
  });

  it('handles empty string', () => {
    expect(toSlug('')).toBe('');
  });
});

describe('isValidSlug', () => {
  it('returns true for valid slugs', () => {
    expect(isValidSlug('hello')).toBe(true);
    expect(isValidSlug('senior-frontend-engineer')).toBe(true);
  });

  it('returns false for slugs with uppercase', () => {
    expect(isValidSlug('Hello-World')).toBe(false);
  });

  it('returns false for trailing hyphens', () => {
    expect(isValidSlug('hello-')).toBe(false);
  });

  it('returns false for empty string', () => {
    expect(isValidSlug('')).toBe(false);
  });
});

describe('slugToTitle', () => {
  it('converts slug to title case', () => {
    expect(slugToTitle('senior-engineer')).toBe('Senior Engineer');
  });
});

describe('careerUrl', () => {
  it('builds the public career URL', () => {
    expect(careerUrl('acme', 'senior-dev', 'https://app.rtylr.com')).toBe(
      'https://app.rtylr.com/acme/careers/senior-dev'
    );
  });

  it('works without a base', () => {
    expect(careerUrl('acme', 'senior-dev')).toBe('/acme/careers/senior-dev');
  });
});

describe('applyUrl', () => {
  it('builds the application form URL', () => {
    expect(applyUrl('acme', 'senior-dev')).toBe('/acme/careers/senior-dev/apply');
  });
});
