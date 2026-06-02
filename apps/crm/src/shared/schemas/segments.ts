import { z } from 'zod';

/**
 * Segment wire shape mirrors the backend segments module
 * (onion segments/internal/routes/handlers/segment_handlers.go → SegmentPayload).
 * The backend returns: id, ownerId, name, description, filters (raw JSON),
 * customerCount, createdAt, updatedAt. The @voxire/api response interceptor
 * converts snake_case → camelCase.
 *
 * The backend does NOT send `color`, `icon`, `rules`, `trend`, or `trendUp` —
 * those are presentation concerns generated/derived client-side. They are kept
 * as optional / defaulted here so list responses parse cleanly while the
 * existing segment screens keep their typed access. `rules` is a client-only
 * display projection of `filters`.
 */
export const segmentSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  ownerId: z.number().optional(),
  name: z.string(),
  description: z.string().default(''),
  filters: z.unknown().optional(),
  customerCount: z.number().default(0),
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
