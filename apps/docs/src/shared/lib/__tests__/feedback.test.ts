import { describe, it, expect } from 'vitest';
import { helpfulPercent, feedbackForSlug, needsImprovement, groupBySlug } from '../feedback';

const entries = [
  { slug: 'guide', rating: 'helpful' as const },
  { slug: 'guide', rating: 'helpful' as const },
  { slug: 'guide', rating: 'not_helpful' as const },
  { slug: 'api', rating: 'helpful' as const },
  { slug: 'api', rating: 'not_helpful' as const },
];

describe('helpfulPercent', () => {
  it('calculates percent correctly', () => {
    const guideEntries = entries.filter((e) => e.slug === 'guide');
    expect(helpfulPercent(guideEntries)).toBe(67);
  });

  it('returns 0 for empty entries', () => {
    expect(helpfulPercent([])).toBe(0);
  });
});

describe('feedbackForSlug', () => {
  it('filters by slug', () => {
    expect(feedbackForSlug(entries, 'guide').length).toBe(3);
  });

  it('returns empty for missing slug', () => {
    expect(feedbackForSlug(entries, 'missing')).toEqual([]);
  });
});

describe('needsImprovement', () => {
  it('returns false for fewer than 5 entries', () => {
    expect(needsImprovement([{ slug: 'x', rating: 'not_helpful' }])).toBe(false);
  });

  it('returns true when helpful percent is below 60', () => {
    const bad = Array(6).fill({ slug: 'x', rating: 'not_helpful' as const });
    expect(needsImprovement(bad)).toBe(true);
  });

  it('returns false for high-quality pages', () => {
    const good = Array(6).fill({ slug: 'x', rating: 'helpful' as const });
    expect(needsImprovement(good)).toBe(false);
  });
});

describe('groupBySlug', () => {
  it('groups entries by slug', () => {
    const grouped = groupBySlug(entries);
    expect(grouped['guide'].length).toBe(3);
    expect(grouped['api'].length).toBe(2);
  });

  it('returns empty object for no entries', () => {
    expect(groupBySlug([])).toEqual({});
  });
});
