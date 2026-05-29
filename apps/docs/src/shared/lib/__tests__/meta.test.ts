import { describe, it, expect } from 'vitest';
import { pageTitle, truncateDescription, canonicalUrl, normaliseTags, sharesTags } from '../meta';

describe('pageTitle', () => {
  it('appends site name', () => {
    expect(pageTitle('Getting Started')).toBe('Getting Started | rtylr docs');
  });

  it('returns just site name when title is empty', () => {
    expect(pageTitle('')).toBe('rtylr docs');
  });
});

describe('truncateDescription', () => {
  it('returns unchanged when under limit', () => {
    expect(truncateDescription('short')).toBe('short');
  });

  it('truncates long descriptions with ellipsis', () => {
    const long = 'a'.repeat(200);
    const result = truncateDescription(long);
    expect(result.length).toBeLessThanOrEqual(160);
    expect(result.endsWith('...')).toBe(true);
  });
});

describe('canonicalUrl', () => {
  it('prepends base URL', () => {
    expect(canonicalUrl('/getting-started')).toBe('https://docs.rtylr.com/getting-started');
  });

  it('adds leading slash when missing', () => {
    expect(canonicalUrl('guide')).toBe('https://docs.rtylr.com/guide');
  });

  it('accepts custom base', () => {
    expect(canonicalUrl('/p', 'https://example.com')).toBe('https://example.com/p');
  });
});

describe('normaliseTags', () => {
  it('lowercases and trims', () => {
    expect(normaliseTags(['  API  ', 'REST'])).toEqual(['api', 'rest']);
  });

  it('deduplicates', () => {
    expect(normaliseTags(['a', 'A', 'a'])).toEqual(['a']);
  });

  it('sorts alphabetically', () => {
    expect(normaliseTags(['z', 'a', 'm'])).toEqual(['a', 'm', 'z']);
  });
});

describe('sharesTags', () => {
  it('returns true for common tags', () => {
    expect(sharesTags(['api', 'rest'], ['REST', 'guide'])).toBe(true);
  });

  it('returns false with no overlap', () => {
    expect(sharesTags(['api'], ['guide'])).toBe(false);
  });
});
