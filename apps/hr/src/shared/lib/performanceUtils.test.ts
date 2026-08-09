import { describe, it, expect } from 'vitest';
import {
  reviewStatusLabel,
  reviewStatusColor,
  ratingLabel,
  averageRating,
  isReviewEditable,
  isReviewFinalized,
  improvementPercent,
  normalizeScore,
} from './performanceUtils';

describe('reviewStatusLabel', () => {
  it('returns Draft for Draft', () => {
    expect(reviewStatusLabel('Draft')).toBe('Draft');
  });

  it('returns In Progress for In Progress', () => {
    expect(reviewStatusLabel('In Progress')).toBe('In Progress');
  });

  it('returns Completed for Completed', () => {
    expect(reviewStatusLabel('Completed')).toBe('Completed');
  });

  it('returns Acknowledged for Acknowledged', () => {
    expect(reviewStatusLabel('Acknowledged')).toBe('Acknowledged');
  });
});

describe('reviewStatusColor', () => {
  it('returns gray for Draft', () => {
    expect(reviewStatusColor('Draft')).toBe('gray');
  });

  it('returns blue for In Progress', () => {
    expect(reviewStatusColor('In Progress')).toBe('blue');
  });

  it('returns green for Completed', () => {
    expect(reviewStatusColor('Completed')).toBe('green');
  });

  it('returns purple for Acknowledged', () => {
    expect(reviewStatusColor('Acknowledged')).toBe('purple');
  });
});

describe('ratingLabel', () => {
  it('returns Unsatisfactory for 1', () => {
    expect(ratingLabel(1)).toBe('Unsatisfactory');
  });

  it('returns Needs Improvement for 2', () => {
    expect(ratingLabel(2)).toBe('Needs Improvement');
  });

  it('returns Meets Expectations for 3', () => {
    expect(ratingLabel(3)).toBe('Meets Expectations');
  });

  it('returns Exceeds Expectations for 4', () => {
    expect(ratingLabel(4)).toBe('Exceeds Expectations');
  });

  it('returns Outstanding for 5', () => {
    expect(ratingLabel(5)).toBe('Outstanding');
  });
});

describe('averageRating', () => {
  it('returns 0 for empty array', () => {
    expect(averageRating([])).toBe(0);
  });

  it('returns the single value for a one-element array', () => {
    expect(averageRating([4])).toBe(4);
  });

  it('averages two scores', () => {
    expect(averageRating([3, 5])).toBe(4);
  });

  it('rounds to two decimal places', () => {
    expect(averageRating([1, 2, 3])).toBe(2);
  });

  it('handles fractional averages', () => {
    expect(averageRating([3, 4])).toBe(3.5);
  });
});

describe('isReviewEditable', () => {
  it('returns true for Draft', () => {
    expect(isReviewEditable('Draft')).toBe(true);
  });

  it('returns true for In Progress', () => {
    expect(isReviewEditable('In Progress')).toBe(true);
  });

  it('returns false for Completed', () => {
    expect(isReviewEditable('Completed')).toBe(false);
  });

  it('returns false for Acknowledged', () => {
    expect(isReviewEditable('Acknowledged')).toBe(false);
  });
});

describe('isReviewFinalized', () => {
  it('returns true for Completed', () => {
    expect(isReviewFinalized('Completed')).toBe(true);
  });

  it('returns true for Acknowledged', () => {
    expect(isReviewFinalized('Acknowledged')).toBe(true);
  });

  it('returns false for Draft', () => {
    expect(isReviewFinalized('Draft')).toBe(false);
  });

  it('returns false for In Progress', () => {
    expect(isReviewFinalized('In Progress')).toBe(false);
  });
});

describe('improvementPercent', () => {
  it('returns null when previous is 0', () => {
    expect(improvementPercent(0, 5)).toBeNull();
  });

  it('returns 0 when scores are equal', () => {
    expect(improvementPercent(4, 4)).toBe(0);
  });

  it('returns positive for improvement', () => {
    expect(improvementPercent(4, 5)).toBe(25);
  });

  it('returns negative for regression', () => {
    expect(improvementPercent(5, 4)).toBe(-20);
  });
});

describe('normalizeScore', () => {
  it('returns 0 when maxScore is 0', () => {
    expect(normalizeScore(5, 0)).toBe(0);
  });

  it('returns 100 when score equals maxScore', () => {
    expect(normalizeScore(10, 10)).toBe(100);
  });

  it('returns 50 for half the max', () => {
    expect(normalizeScore(5, 10)).toBe(50);
  });

  it('caps at 100 when score exceeds maxScore', () => {
    expect(normalizeScore(15, 10)).toBe(100);
  });

  it('returns 0 for a score of 0', () => {
    expect(normalizeScore(0, 10)).toBe(0);
  });
});
