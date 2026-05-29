import { describe, it, expect } from 'vitest';
import { getRelated, getRandomDocs } from '../related';

const docs = [
  { slug: 'a', title: 'A', tags: ['api', 'rest'] },
  { slug: 'b', title: 'B', tags: ['api', 'graphql'] },
  { slug: 'c', title: 'C', tags: ['guide', 'tutorial'] },
  { slug: 'd', title: 'D', tags: ['rest', 'guide'] },
];

describe('getRelated', () => {
  it('returns docs with shared tags', () => {
    const related = getRelated(docs[0], docs);
    expect(related.length).toBeGreaterThan(0);
    expect(related.every((d) => d.slug !== 'a')).toBe(true);
  });

  it('respects maxResults', () => {
    const related = getRelated(docs[0], docs, 1);
    expect(related.length).toBeLessThanOrEqual(1);
  });

  it('excludes docs with no shared tags', () => {
    const related = getRelated(docs[0], docs);
    expect(related.find((d) => d.slug === 'c')).toBeUndefined();
  });

  it('returns empty when current has no tags', () => {
    const current = { slug: 'x', tags: [] };
    expect(getRelated(current, docs)).toEqual([]);
  });
});

describe('getRandomDocs', () => {
  it('excludes the current slug', () => {
    const result = getRandomDocs(docs, 'a', 10);
    expect(result.find((d) => d.slug === 'a')).toBeUndefined();
  });

  it('returns at most count items', () => {
    const result = getRandomDocs(docs, 'a', 2);
    expect(result.length).toBeLessThanOrEqual(2);
  });

  it('returns all remaining when count exceeds pool', () => {
    const result = getRandomDocs(docs, 'a', 100);
    expect(result.length).toBe(3);
  });
});
