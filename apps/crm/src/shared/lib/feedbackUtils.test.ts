import { describe, it, expect } from 'vitest';
import {
  averageNps,
  averageRating,
  npsScore,
  groupBySource,
  filterByRatingRange,
  sortByRating,
} from './feedbackUtils';
import type { FeedbackItem } from '../schemas/feedback';

const makeItem = (overrides: Partial<FeedbackItem> = {}): FeedbackItem => ({
  id: 1,
  customer: 'Alice',
  rating: 5,
  nps: 10,
  comment: 'Great!',
  createdAt: '2024-05-01',
  source: 'survey',
  ...overrides,
});

const items: FeedbackItem[] = [
  makeItem({ id: 1, rating: 5, nps: 10, source: 'survey' }),
  makeItem({ id: 2, rating: 3, nps: 7, source: 'pos' }),
  makeItem({ id: 3, rating: 2, nps: 5, source: 'survey' }),
  makeItem({ id: 4, rating: 4, nps: 9, source: 'app' }),
];

describe('averageNps', () => {
  it('returns the correct average NPS', () => {
    expect(averageNps(items)).toBeCloseTo((10 + 7 + 5 + 9) / 4);
  });

  it('returns 0 for empty array', () => {
    expect(averageNps([])).toBe(0);
  });
});

describe('averageRating', () => {
  it('returns the correct average rating', () => {
    expect(averageRating(items)).toBeCloseTo((5 + 3 + 2 + 4) / 4);
  });

  it('returns 0 for empty array', () => {
    expect(averageRating([])).toBe(0);
  });
});

describe('npsScore', () => {
  it('calculates NPS as (promoters - detractors) / total * 100', () => {
    // promoters (nps >= 9): items 1, 4 => 2
    // detractors (nps <= 6): item 3 => 1
    // passives: item 2 (nps=7) => 1
    // NPS = (2-1)/4 * 100 = 25
    expect(npsScore(items)).toBeCloseTo(25);
  });

  it('returns 0 for empty array', () => {
    expect(npsScore([])).toBe(0);
  });

  it('returns 100 when all are promoters', () => {
    const allPromoters = [makeItem({ nps: 9 }), makeItem({ nps: 10 })];
    expect(npsScore(allPromoters)).toBe(100);
  });
});

describe('groupBySource', () => {
  it('groups items by their source', () => {
    const groups = groupBySource(items);
    expect(groups['survey']).toHaveLength(2);
    expect(groups['pos']).toHaveLength(1);
    expect(groups['app']).toHaveLength(1);
  });
});

describe('filterByRatingRange', () => {
  it('returns items within the rating range', () => {
    const result = filterByRatingRange(items, 4, 5);
    expect(result).toHaveLength(2);
    expect(result.every((i) => i.rating >= 4 && i.rating <= 5)).toBe(true);
  });

  it('returns empty array when no items match', () => {
    expect(filterByRatingRange(items, 1, 1)).toHaveLength(0);
  });
});

describe('sortByRating', () => {
  it('sorts items by rating, highest first', () => {
    const sorted = sortByRating(items);
    expect(sorted[0].rating).toBe(5);
    expect(sorted[sorted.length - 1].rating).toBe(2);
  });

  it('does not mutate the original array', () => {
    const original = [...items];
    sortByRating(items);
    expect(items[0]).toEqual(original[0]);
  });
});
