import { z } from 'zod';

export const settingsSchema = z.object({
  defaultConsent: z.boolean().default(false),
  timezone: z.string().default(''),
  fromEmail: z.string().default(''),
  fromName: z.string().default(''),
  smsSenderId: z.string().default(''),
  // NOTE: the backend GET /crm/settings does not currently return these three.
  // Defaulted so the page loads; they need wiring into SettingsPayload +
  // UpdateSettingsRequest (or removal from the form) to actually persist.
  businessName: z.string().default(''),
  atRiskDays: z.number().default(0),
  currency: z.string().default(''),
});

export type Settings = z.infer<typeof settingsSchema>;
