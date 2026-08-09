import { describe, it, expect } from 'vitest';
import {
  avgRating,
  recommendCount,
  recommendRate,
  ratingLabel,
  hasPendingFeedback,
  topRatedFeedback,
} from './feedback';
import type { InterviewFeedback } from './feedback';

function makeFeedback(rating: 1 | 2 | 3 | 4 | 5, recommend: boolean): InterviewFeedback {
  return { interviewerId: 'i1', rating, recommend, notes: '' };
}

const feedbacks = [makeFeedback(5, true), makeFeedback(4, true), makeFeedback(2, false)];

describe('avgRating', () => {
  it('calculates average rating', () => {
    expect(avgRating(feedbacks)).toBeCloseTo(3.7, 0);
  });

  it('returns 0 for empty feedbacks', () => {
    expect(avgRating([])).toBe(0);
  });
});

describe('recommendCount', () => {
  it('counts recommending interviewers', () => {
    expect(recommendCount(feedbacks)).toBe(2);
  });

  it('returns 0 when none recommend', () => {
    expect(recommendCount([makeFeedback(1, false)])).toBe(0);
  });
});

describe('recommendRate', () => {
  it('returns percentage who recommend', () => {
    expect(recommendRate(feedbacks)).toBeCloseTo(66.7, 0);
  });

  it('returns 0 for empty feedbacks', () => {
    expect(recommendRate([])).toBe(0);
  });
});

describe('ratingLabel', () => {
  it('returns correct label for rating 1', () => {
    expect(ratingLabel(1)).toBe('Strong No');
  });

  it('returns correct label for rating 5', () => {
    expect(ratingLabel(5)).toBe('Strong Yes');
  });

  it('returns Neutral for rating 3', () => {
    expect(ratingLabel(3)).toBe('Neutral');
  });
});

describe('hasPendingFeedback', () => {
  it('returns true when fewer than expected submitted', () => {
    expect(hasPendingFeedback(feedbacks, 5)).toBe(true);
  });

  it('returns false when all submitted', () => {
    expect(hasPendingFeedback(feedbacks, 3)).toBe(false);
  });
});

describe('topRatedFeedback', () => {
  it('returns feedbacks at or above threshold', () => {
    expect(topRatedFeedback(feedbacks, 4)).toHaveLength(2);
  });

  it('returns empty when none qualify', () => {
    expect(topRatedFeedback(feedbacks, 5 as const)).toHaveLength(1);
  });
});
