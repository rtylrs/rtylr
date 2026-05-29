import type { FeedbackItem } from '../schemas/feedback';

/** Returns the average NPS score across all feedback items. */
export function averageNps(items: FeedbackItem[]): number {
  if (items.length === 0) return 0;
  return items.reduce((sum, item) => sum + item.nps, 0) / items.length;
}

/** Returns the average rating across all feedback items. */
export function averageRating(items: FeedbackItem[]): number {
  if (items.length === 0) return 0;
  return items.reduce((sum, item) => sum + item.rating, 0) / items.length;
}

/** Returns the NPS score: percentage of promoters minus percentage of detractors. */
export function npsScore(items: FeedbackItem[]): number {
  if (items.length === 0) return 0;
  const promoters = items.filter((i) => i.nps >= 9).length;
  const detractors = items.filter((i) => i.nps <= 6).length;
  return ((promoters - detractors) / items.length) * 100;
}

/** Groups feedback items by their source. */
export function groupBySource(items: FeedbackItem[]): Record<string, FeedbackItem[]> {
  return items.reduce<Record<string, FeedbackItem[]>>((acc, item) => {
    const key = item.source;
    if (!acc[key]) acc[key] = [];
    acc[key].push(item);
    return acc;
  }, {});
}

/** Returns feedback items filtered by rating range (inclusive). */
export function filterByRatingRange(
  items: FeedbackItem[],
  min: number,
  max: number
): FeedbackItem[] {
  return items.filter((i) => i.rating >= min && i.rating <= max);
}

/** Returns feedback items sorted by rating, highest first. */
export function sortByRating(items: FeedbackItem[]): FeedbackItem[] {
  return [...items].sort((a, b) => b.rating - a.rating);
}
