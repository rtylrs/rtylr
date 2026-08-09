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
  id: '1',
  customerName: 'Alice',
  type: 'review',
  score: 5,
  comment: 'Great!',
  createdAt: '2024-05-01',
  source: 'survey',
  ...overrides,
});

const reviewItems: FeedbackItem[] = [
  makeItem({ id: '1', type: 'review', score: 5, source: 'survey' }),
  makeItem({ id: '2', type: 'review', score: 3, source: 'pos' }),
  makeItem({ id: '3', type: 'review', score: 2, source: 'survey' }),
  makeItem({ id: '4', type: 'review', score: 4, source: 'app' }),
];

const npsItems: FeedbackItem[] = [
  makeItem({ id: '1', type: 'nps', score: 10, source: 'survey' }),
  makeItem({ id: '2', type: 'nps', score: 7, source: 'pos' }),
  makeItem({ id: '3', type: 'nps', score: 5, source: 'survey' }),
  makeItem({ id: '4', type: 'nps', score: 9, source: 'app' }),
];

describe('averageNps', () => {
  it('returns the correct average across nps-type items', () => {
    expect(averageNps(npsItems)).toBeCloseTo((10 + 7 + 5 + 9) / 4);
  });

  it('ignores review-type items', () => {
    expect(averageNps(reviewItems)).toBe(0);
  });

  it('returns 0 for empty array', () => {
    expect(averageNps([])).toBe(0);
  });
});

describe('averageRating', () => {
  it('returns the correct average across review-type items', () => {
    expect(averageRating(reviewItems)).toBeCloseTo((5 + 3 + 2 + 4) / 4);
  });

  it('ignores nps-type items', () => {
    expect(averageRating(npsItems)).toBe(0);
  });

  it('returns 0 for empty array', () => {
    expect(averageRating([])).toBe(0);
  });
});

describe('npsScore', () => {
  it('calculates NPS as (promoters - detractors) / total * 100 over nps-type items', () => {
    // promoters (score >= 9): items 1, 4 => 2
    // detractors (score <= 6): item 3 => 1
    // passives: item 2 (score=7) => 1
    // NPS = (2-1)/4 * 100 = 25
    expect(npsScore(npsItems)).toBeCloseTo(25);
  });

  it('returns 0 for empty array', () => {
    expect(npsScore([])).toBe(0);
  });

  it('returns 100 when all are promoters', () => {
    const allPromoters = [
      makeItem({ type: 'nps', score: 9 }),
      makeItem({ type: 'nps', score: 10 }),
    ];
    expect(npsScore(allPromoters)).toBe(100);
  });
});

describe('groupBySource', () => {
  it('groups items by their source', () => {
    const groups = groupBySource(reviewItems);
    expect(groups['survey']).toHaveLength(2);
    expect(groups['pos']).toHaveLength(1);
    expect(groups['app']).toHaveLength(1);
  });
});

describe('filterByRatingRange', () => {
  it('returns review-type items within the rating range', () => {
    const result = filterByRatingRange(reviewItems, 4, 5);
    expect(result).toHaveLength(2);
    expect(result.every((i) => i.score >= 4 && i.score <= 5)).toBe(true);
  });

  it('returns empty array when no items match', () => {
    expect(filterByRatingRange(reviewItems, 1, 1)).toHaveLength(0);
  });

  it('excludes nps-type items', () => {
    expect(filterByRatingRange(npsItems, 0, 10)).toHaveLength(0);
  });
});

describe('sortByRating', () => {
  it('sorts review-type items by score, highest first', () => {
    const sorted = sortByRating(reviewItems);
    expect(sorted[0].score).toBe(5);
    expect(sorted[sorted.length - 1].score).toBe(2);
  });

  it('does not mutate the original array', () => {
    const original = [...reviewItems];
    sortByRating(reviewItems);
    expect(reviewItems[0]).toEqual(original[0]);
  });
});
