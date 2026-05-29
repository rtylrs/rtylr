import { z } from 'zod';

export const segmentSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  description: z.string(),
  customerCount: z.number(),
  color: z.string(),
  icon: z.string().optional(),
  rules: z.array(z.string()).default([]),
  trend: z.string(),
  trendUp: z.boolean(),
});

export type Segment = z.infer<typeof segmentSchema>;

// -- Constants ----------------------------------------------------------------

export const SEGMENT_COLORS = [
  '#3b82f6',
  '#8b5cf6',
  '#ec4899',
  '#f59e0b',
  '#10b981',
  '#ef4444',
] as const;

// -- Type guards --------------------------------------------------------------

export function isValidSegmentColor(color: string): boolean {
  return (SEGMENT_COLORS as readonly string[]).includes(color);
}

export function isTrendingUp(segment: Segment): boolean {
  return segment.trendUp === true;
}

export const SEGMENT_SIZES = ['small', 'medium', 'large', 'xlarge'] as const;
export type SegmentSize = (typeof SEGMENT_SIZES)[number];

export function getSegmentSize(memberCount: number): SegmentSize {
  if (memberCount < 100) return 'small';
  if (memberCount < 1000) return 'medium';
  if (memberCount < 10000) return 'large';
  return 'xlarge';
}

export function segmentSizeLabel(size: SegmentSize): string {
  const labels: Record<SegmentSize, string> = {
    small: 'Small (<100)',
    medium: 'Medium (100-1K)',
    large: 'Large (1K-10K)',
    xlarge: 'Extra Large (10K+)',
  };
  return labels[size];
}
