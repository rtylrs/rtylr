import { z } from 'zod';

export const hrSettingsSchema = z.object({
  workWeekStart: z.enum(['Monday', 'Sunday']).catch('Monday'),
  overtimeAfterHours: z.number(),
  ptoPolicyDays: z.number(),
  currency: z.enum(['USD', 'EUR', 'GBP']).catch('USD'),
  weeklyHours: z.number(),
});

export type HrSettings = z.infer<typeof hrSettingsSchema>;
