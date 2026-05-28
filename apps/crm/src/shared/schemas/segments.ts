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
