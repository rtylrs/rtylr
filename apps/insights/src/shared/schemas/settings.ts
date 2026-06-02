import { z } from 'zod';

export const insightsSettingsSchema = z.object({
  defaultCurrency: z.string().optional(),
  fiscalYearStart: z.string().optional(),
  timezone: z.string().optional(),
});

export const scheduledReportSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string().optional(),
  templateId: z.string().optional(),
  frequency: z.string().optional(),
  recipients: z.array(z.string()).optional(),
  enabled: z.boolean().optional(),
  lastRunAt: z.string().optional().nullable(),
  nextRunAt: z.string().optional().nullable(),
  data: z.array(z.unknown()).optional(),
});

export type InsightsSettings = z.infer<typeof insightsSettingsSchema>;
export type ScheduledReport = z.infer<typeof scheduledReportSchema>;
