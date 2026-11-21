import { z } from 'zod';

const currencySchema = z.enum(['USD', 'EUR', 'GBP']).catch('USD');

// GET /hr/config returns HRCompanyConfig fields (defaultCurrency,
// overtimeMultiplier, directoryEnabled, etc.) and does not carry
// workWeekStart, weeklyHours, overtimeAfterHours, or ptoPolicyDays.
// Normalize both shapes so HR Settings General loads cleanly.
export const hrSettingsSchema = z.preprocess(
  (input) => {
    const raw = input && typeof input === 'object' ? (input as Record<string, unknown>) : {};
    return {
      workWeekStart: raw.workWeekStart,
      overtimeAfterHours: raw.overtimeAfterHours ?? 40,
      ptoPolicyDays: raw.ptoPolicyDays ?? 20,
      currency: raw.currency ?? raw.defaultCurrency ?? raw.default_currency ?? 'USD',
      weeklyHours: raw.weeklyHours ?? 40,
    };
  },
  z.object({
    workWeekStart: z.enum(['Monday', 'Sunday']).catch('Monday'),
    overtimeAfterHours: z.coerce.number().catch(40),
    ptoPolicyDays: z.coerce.number().catch(20),
    currency: currencySchema,
    weeklyHours: z.coerce.number().catch(40),
  })
);

export type HrSettings = z.infer<typeof hrSettingsSchema>;
