import { z } from 'zod';

export const settingsSchema = z.object({
  defaultConsent: z.boolean(),
  timezone: z.string(),
  fromEmail: z.string(),
  fromName: z.string(),
  smsSenderId: z.string(),
  businessName: z.string(),
  atRiskDays: z.number(),
  currency: z.string(),
});

export type Settings = z.infer<typeof settingsSchema>;
