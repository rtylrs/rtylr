import { describe, it, expect } from 'vitest';
import {
  buildPageTitle,
  buildMetaDescription,
  buildCanonicalUrl,
  buildOgImageUrl,
  buildHreflangLinks,
  pathToBreadcrumbLabel,
  isMetaDescriptionValid,
  isPageTitleValid,
} from './seoUtils';

describe('buildPageTitle', () => {
  it('returns default title when no argument is provided', () => {
    expect(buildPageTitle()).toBe('rtylr - One Suite for Your Entire Business');
  });

  it('appends site name for a custom page title', () => {
    expect(buildPageTitle('POS System')).toBe('POS System | rtylr');
  });

  it('returns the default title unchanged when passed directly', () => {
    const defaultTitle = 'rtylr - One Suite for Your Entire Business';
    expect(buildPageTitle(defaultTitle)).toBe(defaultTitle);
  });
});

describe('buildMetaDescription', () => {
  it('returns the default description when none is provided', () => {
    const result = buildMetaDescription();
    expect(result).toContain('rtylr');
  });

  it('returns the description unchanged when short enough', () => {
    const short = 'A short description.';
    expect(buildMetaDescription(short)).toBe(short);
  });

  it('truncates long descriptions to 155 characters with ellipsis', () => {
    const long = 'a'.repeat(200);
    const result = buildMetaDescription(long);
    expect(result).toHaveLength(155);
    expect(result.endsWith('...')).toBe(true);
  });

  it('does not truncate exactly 155 character descriptions', () => {
    const exact = 'a'.repeat(155);
    expect(buildMetaDescription(exact)).toBe(exact);
  });
});

describe('buildCanonicalUrl', () => {
  it('builds canonical URL with trailing slash', () => {
    expect(buildCanonicalUrl('/pos')).toBe('https://rtylr.com/pos/');
  });

  it('preserves existing trailing slash', () => {
    expect(buildCanonicalUrl('/flow/')).toBe('https://rtylr.com/flow/');
  });

  it('handles root path', () => {
    expect(buildCanonicalUrl('/')).toBe('https://rtylr.com/');
  });

  it('adds leading slash when missing', () => {
    expect(buildCanonicalUrl('crm')).toBe('https://rtylr.com/crm/');
  });
});

describe('buildOgImageUrl', () => {
  it('returns default OG image URL when no path provided', () => {
    expect(buildOgImageUrl()).toBe('https://rtylr.com/og-default.png');
  });

  it('builds absolute URL from relative path', () => {
    expect(buildOgImageUrl('/og/pos.png')).toBe('https://rtylr.com/og/pos.png');
  });

  it('returns absolute URL unchanged', () => {
    const abs = 'https://example.com/image.png';
    expect(buildOgImageUrl(abs)).toBe(abs);
  });

  it('adds leading slash to path without one', () => {
    expect(buildOgImageUrl('og/flow.png')).toBe('https://rtylr.com/og/flow.png');
  });
});

describe('buildHreflangLinks', () => {
  it('returns x-default, en, and ar keys', () => {
    const links = buildHreflangLinks('/pos/');
    expect(Object.keys(links)).toEqual(['x-default', 'en', 'ar']);
  });

  it('sets x-default and en to canonical URL', () => {
    const links = buildHreflangLinks('/pos/');
    expect(links['x-default']).toBe(links['en']);
  });

  it('includes /ar/ prefix for Arabic alternate', () => {
    const links = buildHreflangLinks('/pos/');
    expect(links['ar']).toContain('/ar/');
  });
});

describe('pathToBreadcrumbLabel', () => {
  it('converts a simple path segment to title case', () => {
    expect(pathToBreadcrumbLabel('/flow/')).toBe('Flow');
  });

  it('joins multiple segments with spaces', () => {
    expect(pathToBreadcrumbLabel('/hr/pricing/')).toBe('Hr Pricing');
  });

  it('converts hyphenated segments to title case words', () => {
    expect(pathToBreadcrumbLabel('/point-of-sale/')).toBe('Point Of Sale');
  });

  it('returns Home for root path', () => {
    expect(pathToBreadcrumbLabel('/')).toBe('Home');
  });
});

describe('isMetaDescriptionValid', () => {
  it('returns true for valid length description', () => {
    expect(isMetaDescriptionValid('a'.repeat(100))).toBe(true);
  });

  it('returns false for too short description', () => {
    expect(isMetaDescriptionValid('short')).toBe(false);
  });

  it('returns false for too long description', () => {
    expect(isMetaDescriptionValid('a'.repeat(200))).toBe(false);
  });

  it('returns true for exactly 50 characters', () => {
    expect(isMetaDescriptionValid('a'.repeat(50))).toBe(true);
  });

  it('returns true for exactly 160 characters', () => {
    expect(isMetaDescriptionValid('a'.repeat(160))).toBe(true);
  });
});

describe('isPageTitleValid', () => {
  it('returns true for a valid title', () => {
    expect(isPageTitleValid('POS System | rtylr')).toBe(true);
  });

  it('returns false for title shorter than 10 characters', () => {
    expect(isPageTitleValid('Short')).toBe(false);
  });

  it('returns false for title longer than 70 characters', () => {
    expect(isPageTitleValid('a'.repeat(71))).toBe(false);
  });

  it('returns true for exactly 10 characters', () => {
    expect(isPageTitleValid('a'.repeat(10))).toBe(true);
  });

  it('returns true for exactly 70 characters', () => {
    expect(isPageTitleValid('a'.repeat(70))).toBe(true);
  });
});
