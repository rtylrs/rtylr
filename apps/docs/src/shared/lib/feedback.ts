/** Docs page feedback helpers. */

export type FeedbackRating = 'helpful' | 'not_helpful';

export interface FeedbackEntry {
  slug: string;
  rating: FeedbackRating;
  timestamp?: number;
}

/**
 * Returns the percentage of helpful feedback from an array of entries.
 * Returns 0 if there are no entries.
 */
export function helpfulPercent(entries: FeedbackEntry[]): number {
  if (entries.length === 0) return 0;
  const helpful = entries.filter((e) => e.rating === 'helpful').length;
  return Math.round((helpful / entries.length) * 100);
}

/**
 * Returns feedback entries for a specific page slug.
 */
export function feedbackForSlug(entries: FeedbackEntry[], slug: string): FeedbackEntry[] {
  return entries.filter((e) => e.slug === slug);
}

/**
 * Returns true if the page needs improvement (less than 60% helpful).
 */
export function needsImprovement(entries: FeedbackEntry[]): boolean {
  if (entries.length < 5) return false;
  return helpfulPercent(entries) < 60;
}

/**
 * Groups feedback entries by slug.
 */
export function groupBySlug(entries: FeedbackEntry[]): Record<string, FeedbackEntry[]> {
  return entries.reduce(
    (acc, entry) => {
      if (!acc[entry.slug]) acc[entry.slug] = [];
      acc[entry.slug].push(entry);
      return acc;
    },
    {} as Record<string, FeedbackEntry[]>
  );
}
