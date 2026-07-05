import type { FeedbackItem } from '../schemas/feedback';

/** Returns the average score across nps-type feedback items. */
export function averageNps(items: FeedbackItem[]): number {
  const npsItems = items.filter((i) => i.type === 'nps');
  if (npsItems.length === 0) return 0;
  return npsItems.reduce((sum, item) => sum + item.score, 0) / npsItems.length;
}

/** Returns the average star rating across review-type feedback items. */
export function averageRating(items: FeedbackItem[]): number {
  const reviewItems = items.filter((i) => i.type === 'review');
  if (reviewItems.length === 0) return 0;
  return reviewItems.reduce((sum, item) => sum + item.score, 0) / reviewItems.length;
}

/** Returns the NPS score: percentage of promoters minus percentage of detractors, over nps-type items. */
export function npsScore(items: FeedbackItem[]): number {
  const npsItems = items.filter((i) => i.type === 'nps');
  if (npsItems.length === 0) return 0;
  const promoters = npsItems.filter((i) => i.score >= 9).length;
  const detractors = npsItems.filter((i) => i.score <= 6).length;
  return ((promoters - detractors) / npsItems.length) * 100;
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

/** Returns review-type feedback items with a star rating within range (inclusive). */
export function filterByRatingRange(
  items: FeedbackItem[],
  min: number,
  max: number
): FeedbackItem[] {
  return items.filter((i) => i.type === 'review' && i.score >= min && i.score <= max);
}

/** Returns review-type feedback items sorted by star rating, highest first. */
export function sortByRating(items: FeedbackItem[]): FeedbackItem[] {
  return items.filter((i) => i.type === 'review').sort((a, b) => b.score - a.score);
}
