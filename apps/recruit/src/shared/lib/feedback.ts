// Interview feedback helpers for the recruit admin panel.

export type FeedbackRating = 1 | 2 | 3 | 4 | 5;

export interface InterviewFeedback {
  interviewerId: string;
  rating: FeedbackRating;
  recommend: boolean;
  notes: string;
}

/**
 * avgRating returns the average rating across all feedback entries.
 */
export function avgRating(feedbacks: InterviewFeedback[]): number {
  if (feedbacks.length === 0) return 0;
  const sum = feedbacks.reduce((acc, f) => acc + f.rating, 0);
  return Math.round((sum / feedbacks.length) * 10) / 10;
}

/**
 * recommendCount returns the number of interviewers who recommend the candidate.
 */
export function recommendCount(feedbacks: InterviewFeedback[]): number {
  return feedbacks.filter((f) => f.recommend).length;
}

/**
 * recommendRate returns the percentage of interviewers who recommend the candidate.
 */
export function recommendRate(feedbacks: InterviewFeedback[]): number {
  if (feedbacks.length === 0) return 0;
  return Math.round((recommendCount(feedbacks) / feedbacks.length) * 1000) / 10;
}

/**
 * ratingLabel returns a human-readable label for a numeric rating.
 */
export function ratingLabel(rating: FeedbackRating): string {
  const labels: Record<FeedbackRating, string> = {
    1: 'Strong No',
    2: 'No',
    3: 'Neutral',
    4: 'Yes',
    5: 'Strong Yes',
  };
  return labels[rating];
}

/**
 * hasPendingFeedback returns true when fewer than expected interviewers have submitted feedback.
 */
export function hasPendingFeedback(feedbacks: InterviewFeedback[], expectedCount: number): boolean {
  return feedbacks.length < expectedCount;
}

/**
 * topRatedFeedback returns feedback entries at or above the given rating threshold.
 */
export function topRatedFeedback(
  feedbacks: InterviewFeedback[],
  threshold: FeedbackRating
): InterviewFeedback[] {
  return feedbacks.filter((f) => f.rating >= threshold);
}
