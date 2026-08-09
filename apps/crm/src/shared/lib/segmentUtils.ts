import type { Segment } from '../schemas/segments';

/** Sorts segments by customer count, largest first. */
export function sortBySize(segments: Segment[]): Segment[] {
  return [...segments].sort((a, b) => b.customerCount - a.customerCount);
}

/** Returns segments whose trend is upward. */
export function getTrendingSegments(segments: Segment[]): Segment[] {
  return segments.filter((s) => s.trendUp);
}

/** Returns the total customer count across all segments. */
export function totalCustomers(segments: Segment[]): number {
  return segments.reduce((sum, s) => sum + s.customerCount, 0);
}

/** Returns the segment with the highest customer count, or null for empty input. */
export function largestSegment(segments: Segment[]): Segment | null {
  if (segments.length === 0) return null;
  return segments.reduce((max, s) => (s.customerCount > max.customerCount ? s : max));
}

/**
 * Returns true if a segment with the given name exists (case-insensitive).
 */
export function segmentNameExists(segments: Segment[], name: string): boolean {
  const lower = name.toLowerCase();
  return segments.some((s) => s.name.toLowerCase() === lower);
}

/** Returns only segments that have at least one rule defined. */
export function getSegmentsWithRules(segments: Segment[]): Segment[] {
  return segments.filter((s) => s.rules.length > 0);
}
