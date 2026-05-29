// Customer segment helpers for the insights analytics feature.

export interface CustomerSegment {
  id: string;
  name: string;
  count: number;
  revenue: number;
}

/**
 * totalSegmentRevenue returns the sum of revenue across all segments.
 */
export function totalSegmentRevenue(segments: CustomerSegment[]): number {
  return segments.reduce((sum, s) => sum + s.revenue, 0);
}

/**
 * totalSegmentCount returns the sum of customer counts across all segments.
 */
export function totalSegmentCount(segments: CustomerSegment[]): number {
  return segments.reduce((sum, s) => sum + s.count, 0);
}

/**
 * segmentRevenueShare returns the revenue share percentage of a segment.
 */
export function segmentRevenueShare(segment: CustomerSegment, segments: CustomerSegment[]): number {
  const total = totalSegmentRevenue(segments);
  if (total === 0) return 0;
  return Math.round((segment.revenue / total) * 1000) / 10;
}

/**
 * topSegmentByRevenue returns the segment with the highest revenue.
 */
export function topSegmentByRevenue(segments: CustomerSegment[]): CustomerSegment | undefined {
  if (segments.length === 0) return undefined;
  return segments.reduce((best, s) => (s.revenue > best.revenue ? s : best));
}

/**
 * sortSegmentsByRevenue returns segments sorted by revenue descending.
 */
export function sortSegmentsByRevenue(segments: CustomerSegment[]): CustomerSegment[] {
  return [...segments].sort((a, b) => b.revenue - a.revenue);
}

/**
 * avgRevenuePerCustomer returns the average revenue per customer across all segments.
 */
export function avgRevenuePerCustomer(segments: CustomerSegment[]): number {
  const count = totalSegmentCount(segments);
  if (count === 0) return 0;
  return Math.round((totalSegmentRevenue(segments) / count) * 100) / 100;
}
