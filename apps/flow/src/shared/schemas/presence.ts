import { z } from 'zod';

/**
 * `/flow/presence/today` is an ad hoc REST endpoint - presence has no
 * onekit contract, since it aggregates hr_core, hr_leave, and Flow's own
 * work-day data server-side rather than being any one module's concern.
 */
export const presenceEntrySchema = z.object({
  employee_id: z.union([z.string(), z.number()]).transform(String),
  user_id: z.union([z.string(), z.number()]).transform(String),
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  status: z.string(),
});

export const presenceResponseSchema = z.object({
  presence: z.array(presenceEntrySchema),
  date: z.string(),
});

export type PresenceEntry = z.infer<typeof presenceEntrySchema>;
