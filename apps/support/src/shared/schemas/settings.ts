import { z } from 'zod';

export const supportSettingsSchema = z.object({
  autoAssign: z.boolean().optional(),
  roundRobin: z.boolean().optional(),
  slaEnabled: z.boolean().optional(),
  slaFirstResponseMinutes: z.number().nullable().optional(),
  slaResolutionMinutes: z.number().nullable().optional(),
  businessHoursStart: z.string().optional(),
  businessHoursEnd: z.string().optional(),
  businessDays: z.string().optional(),
  emailNotifications: z.boolean().optional(),
  slackNotifications: z.boolean().optional(),
});

export type SupportSettings = z.infer<typeof supportSettingsSchema>;
