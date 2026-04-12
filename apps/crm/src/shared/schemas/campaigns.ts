import { z } from 'zod';

export const campaignSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  description: z.string(),
  channel: z.string(),
  status: z.string(),
  audience: z.number(),
  sent: z.number(),
  openRate: z.string(),
  clickRate: z.string(),
  revenue: z.string(),
  updatedAt: z.string(),
  subject: z.string().nullable().optional(),
  body: z.string().nullable().optional(),
  segmentId: z.number().nullable().optional(),
  scheduledFor: z.string().nullable().optional(),
});

export const campaignStatsSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  sent: z.number(),
  delivered: z.number(),
  opened: z.number(),
  clicked: z.number(),
  unsubscribed: z.number(),
  bounced: z.number(),
  revenue: z.number(),
  timeline: z
    .array(
      z.object({
        day: z.string(),
        opens: z.number(),
        clicks: z.number(),
      })
    )
    .default([]),
});

export type Campaign = z.infer<typeof campaignSchema>;
export type CampaignStats = z.infer<typeof campaignStatsSchema>;
