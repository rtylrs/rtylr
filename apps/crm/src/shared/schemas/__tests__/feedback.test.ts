import { describe, it, expect } from 'vitest';
import {
  feedbackItemSchema,
  FEEDBACK_SOURCES,
  isFeedbackSource,
  isPositiveFeedback,
  isNegativeFeedback,
  isPromoter,
  isDetractor,
} from '../feedback';

const baseFeedback = {
  id: 1,
  customer: 'Alice Johnson',
  rating: 5,
  nps: 10,
  comment: 'Great experience!',
  createdAt: '2024-05-01',
  source: 'survey',
};

describe('feedbackItemSchema', () => {
  it('parses a valid feedback item', () => {
    const result = feedbackItemSchema.parse(baseFeedback);
    expect(result.id).toBe(1);
    expect(result.customer).toBe('Alice Johnson');
    expect(result.rating).toBe(5);
  });

  it('coerces string id to number', () => {
    const result = feedbackItemSchema.parse({ ...baseFeedback, id: '12' });
    expect(result.id).toBe(12);
  });
});

describe('FEEDBACK_SOURCES', () => {
  it('contains all expected sources', () => {
    expect(FEEDBACK_SOURCES).toContain('survey');
    expect(FEEDBACK_SOURCES).toContain('review');
    expect(FEEDBACK_SOURCES).toContain('pos');
    expect(FEEDBACK_SOURCES).toContain('email');
    expect(FEEDBACK_SOURCES).toContain('app');
  });

  it('has exactly 5 sources', () => {
    expect(FEEDBACK_SOURCES).toHaveLength(5);
  });
});

describe('isFeedbackSource', () => {
  it('returns true for valid sources', () => {
    expect(isFeedbackSource('survey')).toBe(true);
    expect(isFeedbackSource('pos')).toBe(true);
    expect(isFeedbackSource('app')).toBe(true);
  });

  it('returns false for invalid sources', () => {
    expect(isFeedbackSource('twitter')).toBe(false);
    expect(isFeedbackSource('')).toBe(false);
  });
});

describe('isPositiveFeedback', () => {
  it('returns true for rating >= 4', () => {
    const item = feedbackItemSchema.parse({ ...baseFeedback, rating: 4 });
    expect(isPositiveFeedback(item)).toBe(true);
    const item5 = feedbackItemSchema.parse({ ...baseFeedback, rating: 5 });
    expect(isPositiveFeedback(item5)).toBe(true);
  });

  it('returns false for rating < 4', () => {
    const item = feedbackItemSchema.parse({ ...baseFeedback, rating: 3 });
    expect(isPositiveFeedback(item)).toBe(false);
  });
});

describe('isNegativeFeedback', () => {
  it('returns true for rating <= 2', () => {
    const item = feedbackItemSchema.parse({ ...baseFeedback, rating: 2 });
    expect(isNegativeFeedback(item)).toBe(true);
    const item1 = feedbackItemSchema.parse({ ...baseFeedback, rating: 1 });
    expect(isNegativeFeedback(item1)).toBe(true);
  });

  it('returns false for rating > 2', () => {
    const item = feedbackItemSchema.parse({ ...baseFeedback, rating: 3 });
    expect(isNegativeFeedback(item)).toBe(false);
  });
});

describe('isPromoter', () => {
  it('returns true for nps >= 9', () => {
    const item = feedbackItemSchema.parse({ ...baseFeedback, nps: 9 });
    expect(isPromoter(item)).toBe(true);
    const item10 = feedbackItemSchema.parse({ ...baseFeedback, nps: 10 });
    expect(isPromoter(item10)).toBe(true);
  });

  it('returns false for nps < 9', () => {
    const item = feedbackItemSchema.parse({ ...baseFeedback, nps: 8 });
    expect(isPromoter(item)).toBe(false);
  });
});

describe('isDetractor', () => {
  it('returns true for nps <= 6', () => {
    const item = feedbackItemSchema.parse({ ...baseFeedback, nps: 6 });
    expect(isDetractor(item)).toBe(true);
    const item0 = feedbackItemSchema.parse({ ...baseFeedback, nps: 0 });
    expect(isDetractor(item0)).toBe(true);
  });

  it('returns false for nps > 6', () => {
    const item = feedbackItemSchema.parse({ ...baseFeedback, nps: 7 });
    expect(isDetractor(item)).toBe(false);
  });
});
