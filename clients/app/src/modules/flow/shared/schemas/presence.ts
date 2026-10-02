import { z } from 'zod';

/**
 * `/flow/presence/today` is an ad hoc REST endpoint - presence has no
 * onekit contract, since it aggregates hr_core, hr_leave, and Flow's own
 * work-day data server-side rather than being any one module's concern.
 */
/**
 * camelCase, not the wire's snake_case. `@voxire/api`'s response interceptor
 * runs `convertKeys`/`snakeToCamel` over every response before the caller sees
 * it, so by the time this schema runs the entry is already
 * `{employeeId, userId, firstName, lastName, status}`. Parsing for snake_case
 * made both id unions fail on `undefined`, the whole parse throw, and
 * `/flow/presence` fall back to its error state on a 200 with a full payload.
 *
 * Only camelCase is accepted. Also allowing snake_case would be accepting a
 * shape this client can never receive, and would hide the same mistake next
 * time.
 */
export const presenceEntrySchema = z.object({
  employeeId: z.union([z.string(), z.number()]).transform(String),
  userId: z.union([z.string(), z.number()]).transform(String),
  firstName: z.string().optional(),
  lastName: z.string().optional(),
  status: z.string(),
});

export const presenceResponseSchema = z.object({
  presence: z.array(presenceEntrySchema),
  date: z.string(),
});

export type PresenceEntry = z.infer<typeof presenceEntrySchema>;
