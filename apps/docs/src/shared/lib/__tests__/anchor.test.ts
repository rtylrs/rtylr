import { describe, it, expect } from 'vitest';
import { headingToAnchor, anchorUrl, extractAnchor, hasAnchor, anchorsMatch } from '../anchor';

describe('headingToAnchor', () => {
  it('converts heading to anchor', () => {
    expect(headingToAnchor('Getting Started')).toBe('getting-started');
  });

  it('removes special characters', () => {
    expect(headingToAnchor('What is OAuth?')).toBe('what-is-oauth');
  });

  it('trims leading and trailing hyphens', () => {
    expect(headingToAnchor('  My Section  ')).toBe('my-section');
  });
});

describe('anchorUrl', () => {
  it('appends hash to page URL', () => {
    expect(anchorUrl('/docs/guide', 'installation')).toBe('/docs/guide#installation');
  });

  it('does not double up hashes', () => {
    expect(anchorUrl('/docs/guide', '#installation')).toBe('/docs/guide#installation');
  });
});

describe('extractAnchor', () => {
  it('extracts the fragment', () => {
    expect(extractAnchor('/docs/guide#installation')).toBe('installation');
  });

  it('returns empty string for URL without fragment', () => {
    expect(extractAnchor('/docs/guide')).toBe('');
  });
});

describe('hasAnchor', () => {
  it('returns true for URLs with fragment', () => {
    expect(hasAnchor('/docs#section')).toBe(true);
  });

  it('returns false for plain URLs', () => {
    expect(hasAnchor('/docs/guide')).toBe(false);
  });
});

describe('anchorsMatch', () => {
  it('matches anchors case-insensitively', () => {
    expect(anchorsMatch('#Installation', 'installation')).toBe(true);
  });

  it('returns false for different anchors', () => {
    expect(anchorsMatch('#install', '#setup')).toBe(false);
  });
});
