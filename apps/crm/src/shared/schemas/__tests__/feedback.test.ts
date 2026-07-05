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

const baseReview = {
  id: '1',
  customerName: 'Alice Johnson',
  type: 'review' as const,
  score: 5,
  comment: 'Great experience!',
  createdAt: '2024-05-01',
  source: 'survey',
};

const baseNps = { ...baseReview, type: 'nps' as const, score: 10 };

describe('feedbackItemSchema', () => {
  it('parses a valid review feedback item', () => {
    const result = feedbackItemSchema.parse(baseReview);
    expect(result.id).toBe('1');
    expect(result.customerName).toBe('Alice Johnson');
    expect(result.type).toBe('review');
    expect(result.score).toBe(5);
  });

  it('parses a valid nps feedback item', () => {
    const result = feedbackItemSchema.parse(baseNps);
    expect(result.type).toBe('nps');
    expect(result.score).toBe(10);
  });

  it('keeps id as a string', () => {
    const result = feedbackItemSchema.parse({ ...baseReview, id: '12' });
    expect(result.id).toBe('12');
  });

  it('rejects an unknown type', () => {
    expect(() => feedbackItemSchema.parse({ ...baseReview, type: 'survey' })).toThrow();
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
  it('returns true for review score >= 4', () => {
    const item = feedbackItemSchema.parse({ ...baseReview, score: 4 });
    expect(isPositiveFeedback(item)).toBe(true);
    const item5 = feedbackItemSchema.parse({ ...baseReview, score: 5 });
    expect(isPositiveFeedback(item5)).toBe(true);
  });

  it('returns false for review score < 4', () => {
    const item = feedbackItemSchema.parse({ ...baseReview, score: 3 });
    expect(isPositiveFeedback(item)).toBe(false);
  });

  it('returns false for an nps item regardless of score', () => {
    const item = feedbackItemSchema.parse({ ...baseNps, score: 10 });
    expect(isPositiveFeedback(item)).toBe(false);
  });
});

describe('isNegativeFeedback', () => {
  it('returns true for review score <= 2', () => {
    const item = feedbackItemSchema.parse({ ...baseReview, score: 2 });
    expect(isNegativeFeedback(item)).toBe(true);
    const item1 = feedbackItemSchema.parse({ ...baseReview, score: 1 });
    expect(isNegativeFeedback(item1)).toBe(true);
  });

  it('returns false for review score > 2', () => {
    const item = feedbackItemSchema.parse({ ...baseReview, score: 3 });
    expect(isNegativeFeedback(item)).toBe(false);
  });

  it('returns false for an nps item regardless of score', () => {
    const item = feedbackItemSchema.parse({ ...baseNps, score: 0 });
    expect(isNegativeFeedback(item)).toBe(false);
  });
});

describe('isPromoter', () => {
  it('returns true for nps score >= 9', () => {
    const item = feedbackItemSchema.parse({ ...baseNps, score: 9 });
    expect(isPromoter(item)).toBe(true);
    const item10 = feedbackItemSchema.parse({ ...baseNps, score: 10 });
    expect(isPromoter(item10)).toBe(true);
  });

  it('returns false for nps score < 9', () => {
    const item = feedbackItemSchema.parse({ ...baseNps, score: 8 });
    expect(isPromoter(item)).toBe(false);
  });

  it('returns false for a review item regardless of score', () => {
    const item = feedbackItemSchema.parse({ ...baseReview, score: 5 });
    expect(isPromoter(item)).toBe(false);
  });
});

describe('isDetractor', () => {
  it('returns true for nps score <= 6', () => {
    const item = feedbackItemSchema.parse({ ...baseNps, score: 6 });
    expect(isDetractor(item)).toBe(true);
    const item0 = feedbackItemSchema.parse({ ...baseNps, score: 0 });
    expect(isDetractor(item0)).toBe(true);
  });

  it('returns false for nps score > 6', () => {
    const item = feedbackItemSchema.parse({ ...baseNps, score: 7 });
    expect(isDetractor(item)).toBe(false);
  });

  it('returns false for a review item regardless of score', () => {
    const item = feedbackItemSchema.parse({ ...baseReview, score: 1 });
    expect(isDetractor(item)).toBe(false);
  });
});
