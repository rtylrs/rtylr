import { describe, it, expect } from 'vitest';
import type { DocContent } from '../../content';
import {
  slugToTitle,
  titleToSlug,
  getSortedSlugs,
  hasSlug,
  bodyLength,
  slugsContaining,
} from '../content';

const makeDoc = (title: string, body: string): DocContent => ({
  title,
  body,
});

const sampleMap: Record<string, DocContent> = {
  'quick-start': makeDoc('Quick Start', '<p>Get started quickly.</p>'),
  'api-overview': makeDoc('API Overview', '<h2>Overview</h2><p>Explore the API.</p>'),
  introduction: makeDoc('Introduction', '<p>Welcome to the docs.</p>'),
};

describe('slugToTitle', () => {
  it('converts a simple slug to title case', () => {
    expect(slugToTitle('quick-start')).toBe('Quick Start');
  });

  it('handles a single word slug', () => {
    expect(slugToTitle('introduction')).toBe('Introduction');
  });

  it('handles multiple words', () => {
    expect(slugToTitle('api-overview-guide')).toBe('Api Overview Guide');
  });
});

describe('titleToSlug', () => {
  it('converts a title to lowercase hyphenated slug', () => {
    expect(titleToSlug('Quick Start')).toBe('quick-start');
  });

  it('strips special characters', () => {
    expect(titleToSlug('Hello, World!')).toBe('hello-world');
  });

  it('collapses multiple spaces', () => {
    expect(titleToSlug('  API   Overview  ')).toBe('api-overview');
  });
});

describe('getSortedSlugs', () => {
  it('returns slugs sorted alphabetically', () => {
    const result = getSortedSlugs(sampleMap);
    expect(result).toEqual(['api-overview', 'introduction', 'quick-start']);
  });

  it('returns empty array for empty map', () => {
    expect(getSortedSlugs({})).toEqual([]);
  });
});

describe('hasSlug', () => {
  it('returns true for existing slug', () => {
    expect(hasSlug(sampleMap, 'quick-start')).toBe(true);
  });

  it('returns false for missing slug', () => {
    expect(hasSlug(sampleMap, 'nonexistent')).toBe(false);
  });
});

describe('bodyLength', () => {
  it('strips HTML tags and returns plain text length', () => {
    const doc = makeDoc('Test', '<p>Hello world</p>');
    expect(bodyLength(doc)).toBe('Hello world'.length);
  });

  it('returns 0 for empty body', () => {
    expect(bodyLength(makeDoc('Empty', ''))).toBe(0);
  });
});

describe('slugsContaining', () => {
  it('finds slugs with keyword in body', () => {
    const result = slugsContaining(sampleMap, 'explore');
    expect(result).toContain('api-overview');
  });

  it('finds slugs with keyword in title', () => {
    const result = slugsContaining(sampleMap, 'introduction');
    expect(result).toContain('introduction');
  });

  it('returns empty array when no match', () => {
    expect(slugsContaining(sampleMap, 'xyzzy')).toEqual([]);
  });

  it('is case-insensitive', () => {
    const result = slugsContaining(sampleMap, 'WELCOME');
    expect(result).toContain('introduction');
  });
});
