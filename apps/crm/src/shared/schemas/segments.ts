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
