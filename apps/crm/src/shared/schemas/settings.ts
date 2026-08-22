import { z } from 'zod';

export const settingsSchema = z.object({
  defaultConsent: z.boolean().default(false),
  timezone: z.string().default(''),
  fromEmail: z.string().default(''),
  fromName: z.string().default(''),
  smsSenderId: z.string().default(''),
});

export type Settings = z.infer<typeof settingsSchema>;

export function normalizeSettings(value: Partial<Settings>): Settings {
  const parsed = settingsSchema.parse(value);
  return {
    ...parsed,
    timezone: parsed.timezone.trim(),
    fromEmail: parsed.fromEmail.trim(),
    fromName: parsed.fromName.trim(),
    smsSenderId: parsed.smsSenderId.trim(),
  };
}
