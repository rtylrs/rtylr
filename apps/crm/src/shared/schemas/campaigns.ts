import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === '') return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : value;
}, z.number());

/**
 * Campaign wire shape mirrors the backend campaigns module
 * (onion campaigns/internal/routes/handlers/campaign_handlers.go → CampaignPayload).
 * The @voxire/api response interceptor converts snake_case → camelCase, so the
 * fields below match the Go json tags after conversion: type, audienceSegmentId,
 * sentCount, openCount, scheduledAt, sentAt, createdAt, updatedAt.
 *
 * The backend does NOT compute openRate / clickRate / revenue — derive those
 * client-side from sentCount / openCount when a display needs them.
 */
export const campaignSchema = z.object({
  id: finiteNumber,
  ownerId: finiteNumber.optional(),
  name: z.string(),
  type: z.string().default(''),
  status: z.string(),
  audienceSegmentId: finiteNumber.nullable().optional(),
  subject: z.string().nullable().optional(),
  body: z.string().nullable().optional(),
  audienceCount: finiteNumber.optional(),
  sentCount: finiteNumber.catch(0),
  openCount: finiteNumber.catch(0),
  scheduledAt: z.string().nullable().optional(),
  sentAt: z.string().nullable().optional(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export const campaignStatsSchema = z.object({
  id: finiteNumber,
  sent: finiteNumber,
  delivered: finiteNumber,
  opened: finiteNumber,
  clicked: finiteNumber,
  unsubscribed: finiteNumber,
  bounced: finiteNumber,
  revenue: finiteNumber,
  timeline: z
    .array(
      z.object({
        day: z.string(),
        opens: finiteNumber,
        clicks: finiteNumber,
      })
    )
    .default([]),
});

export type Campaign = z.infer<typeof campaignSchema>;
export type CampaignStats = z.infer<typeof campaignStatsSchema>;

// -- Derived metrics ----------------------------------------------------------

/** Open rate as a percentage (0–100) derived from sent/open counts. */
export function calculateOpenRate(campaign: Campaign): number {
  if (campaign.sentCount === 0) return 0;
  return (campaign.openCount / campaign.sentCount) * 100;
}

// -- Domain constants ---------------------------------------------------------

export const CAMPAIGN_STATUSES = ['draft', 'scheduled', 'active', 'paused', 'completed'] as const;
export type CampaignStatus = (typeof CAMPAIGN_STATUSES)[number];

export const CAMPAIGN_TYPES = ['email', 'sms', 'push'] as const;
export type CampaignType = (typeof CAMPAIGN_TYPES)[number];

// -- Type guards --------------------------------------------------------------

export function isCampaignStatus(value: string): value is CampaignStatus {
  return (CAMPAIGN_STATUSES as readonly string[]).includes(value);
}

export function isCampaignType(value: string): value is CampaignType {
  return (CAMPAIGN_TYPES as readonly string[]).includes(value);
}

export function isActiveCampaign(campaign: Campaign): boolean {
  return campaign.status === 'active';
}

export function isDraftCampaign(campaign: Campaign): boolean {
  return campaign.status === 'draft';
}

export function isScheduledCampaign(campaign: Campaign): boolean {
  return campaign.status === 'scheduled';
}

export function isCompletedCampaign(campaign: Campaign): boolean {
  return campaign.status === 'completed';
}

export const CAMPAIGN_STATUS_LABELS: Record<CampaignStatus, string> = {
  draft: 'Draft',
  scheduled: 'Scheduled',
  active: 'Active',
  paused: 'Paused',
  completed: 'Completed',
};

export const CAMPAIGN_TYPE_LABELS: Record<CampaignType, string> = {
  email: 'Email',
  sms: 'SMS',
  push: 'Push Notification',
};

export function campaignStatusLabel(status: CampaignStatus): string {
  return CAMPAIGN_STATUS_LABELS[status];
}

export function campaignTypeLabel(type: CampaignType): string {
  return CAMPAIGN_TYPE_LABELS[type];
}
