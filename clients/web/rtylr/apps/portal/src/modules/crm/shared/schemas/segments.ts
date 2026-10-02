import { z } from 'zod';

/**
 * Segment wire shape mirrors crm.segment.v1.Segment. id/ownerId/customerCount
 * are transported as strings (onekit int64-as-string encoding) - kept as
 * strings for id/ownerId (opaque identifiers), transformed to a number for
 * customerCount since the UI does arithmetic/formatting on it.
 *
 * The backend does NOT send `color`, `icon`, `rules`, `trend`, or `trendUp` —
 * those are presentation concerns generated/derived client-side. They are kept
 * as optional / defaulted here so list responses parse cleanly while the
 * existing segment screens keep their typed access. `rules` is a client-only
 * display projection of `filters`.
 */
export const segmentSchema = z.object({
  id: z.string(),
  ownerId: z.string().optional(),
  name: z.string(),
  description: z.string().default(''),
  filters: z.string().default('{}'),
  customerCount: z.union([z.string(), z.number()]).transform(Number).default(0),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
  // Client-only presentation fields — not part of the API contract.
  color: z.string().default('#3b82f6'),
  icon: z.string().optional(),
  rules: z.array(z.string()).default([]),
  trend: z.string().default(''),
  trendUp: z.boolean().default(false),
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
