/**
 * Pure performance review utility helpers for the HR performance module.
 * No external dependencies; all functions operate on plain values.
 */

export type ReviewStatus = 'Draft' | 'In Progress' | 'Completed' | 'Acknowledged';

export type RatingScale = 1 | 2 | 3 | 4 | 5;

/**
 * Returns the display label for a performance review status.
 */
export function reviewStatusLabel(status: ReviewStatus): string {
  const labels: Record<ReviewStatus, string> = {
    Draft: 'Draft',
    'In Progress': 'In Progress',
    Completed: 'Completed',
    Acknowledged: 'Acknowledged',
  };
  return labels[status] ?? status;
}

/**
 * Returns the color token for a performance review status.
 */
export function reviewStatusColor(status: ReviewStatus): string {
  const colors: Record<ReviewStatus, string> = {
    Draft: 'gray',
    'In Progress': 'blue',
    Completed: 'green',
    Acknowledged: 'purple',
  };
  return colors[status] ?? 'gray';
}

/**
 * Returns the human-readable label for a numeric rating (1-5).
 */
export function ratingLabel(rating: RatingScale): string {
  const labels: Record<RatingScale, string> = {
    1: 'Unsatisfactory',
    2: 'Needs Improvement',
    3: 'Meets Expectations',
    4: 'Exceeds Expectations',
    5: 'Outstanding',
  };
  return labels[rating];
}

/**
 * Calculates the average rating from an array of numeric scores.
 * Returns 0 for an empty array.
 * Result is rounded to two decimal places.
 */
export function averageRating(scores: number[]): number {
  if (scores.length === 0) return 0;
  const sum = scores.reduce((a, b) => a + b, 0);
  return Math.round((sum / scores.length) * 100) / 100;
}

/**
 * Returns true when a review is editable (Draft or In Progress).
 */
export function isReviewEditable(status: ReviewStatus): boolean {
  return status === 'Draft' || status === 'In Progress';
}

/**
 * Returns true when a review has been finalized.
 */
export function isReviewFinalized(status: ReviewStatus): boolean {
  return status === 'Completed' || status === 'Acknowledged';
}

/**
 * Calculates performance improvement as a percentage given two period scores.
 * Returns null when the previous score is 0 to avoid division by zero.
 */
export function improvementPercent(previous: number, current: number): number | null {
  if (previous === 0) return null;
  return Math.round(((current - previous) / previous) * 100 * 10) / 10;
}

/**
 * Normalizes a raw score to a 0-100 scale given a maximum possible score.
 */
export function normalizeScore(score: number, maxScore: number): number {
  if (maxScore <= 0) return 0;
  return Math.round((Math.min(score, maxScore) / maxScore) * 100 * 10) / 10;
}
