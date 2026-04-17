import { z } from 'zod';

export const activityEntrySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  actor: z.string(),
  actorId: z.string().optional(),
  action: z.string(),
  target: z.string(),
  app: z.string(),
  ip: z.string().optional(),
  createdAt: z.string(),
  severity: z.enum(['info', 'warning', 'critical']).optional(),
});

export type ActivityEntry = z.infer<typeof activityEntrySchema>;
