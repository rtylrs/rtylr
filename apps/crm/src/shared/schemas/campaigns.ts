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

// -- Domain constants ---------------------------------------------------------

export const CAMPAIGN_STATUSES = ["draft", "scheduled", "active", "paused", "completed"] as const;
export type CampaignStatus = (typeof CAMPAIGN_STATUSES)[number];

export const CAMPAIGN_CHANNELS = ["email", "sms", "push"] as const;
export type CampaignChannel = (typeof CAMPAIGN_CHANNELS)[number];

// -- Type guards --------------------------------------------------------------

export function isCampaignStatus(value: string): value is CampaignStatus {
  return (CAMPAIGN_STATUSES as readonly string[]).includes(value);
}

export function isCampaignChannel(value: string): value is CampaignChannel {
  return (CAMPAIGN_CHANNELS as readonly string[]).includes(value);
}

export function isActiveCampaign(campaign: Campaign): boolean {
  return campaign.status === "active";
}

export function isDraftCampaign(campaign: Campaign): boolean {
  return campaign.status === "draft";
}

export function isScheduledCampaign(campaign: Campaign): boolean {
  return campaign.status === "scheduled";
}

export function isCompletedCampaign(campaign: Campaign): boolean {
  return campaign.status === "completed";
}

export const CAMPAIGN_STATUS_LABELS: Record<CampaignStatus, string> = {
  draft: "Draft",
  scheduled: "Scheduled",
  active: "Active",
  paused: "Paused",
  completed: "Completed",
};

export const CAMPAIGN_CHANNEL_LABELS: Record<CampaignChannel, string> = {
  email: "Email",
  sms: "SMS",
  push: "Push Notification",
};

export function campaignStatusLabel(status: CampaignStatus): string {
  return CAMPAIGN_STATUS_LABELS[status];
}

export function campaignChannelLabel(channel: CampaignChannel): string {
  return CAMPAIGN_CHANNEL_LABELS[channel];
}
