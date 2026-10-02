import { z } from 'zod';

export const settingsSchema = z.object({
  defaultConsent: z.boolean().default(false),
  timezone: z.string().default(''),
  fromEmail: z.string().default(''),
  fromName: z.string().default(''),
  smsSenderId: z.string().default(''),
});

export type Settings = z.infer<typeof settingsSchema>;
