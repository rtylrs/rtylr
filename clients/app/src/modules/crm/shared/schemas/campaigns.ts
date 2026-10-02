import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === '') return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : value;
}, z.number());

/**
 * Campaign reads come from the generated CampaignServiceClient, so the wire
 * carries crm.campaign.v1.Campaign's field names. This schema is declared in
 * those names and then mapped to the ones the UI already uses.
 *
 * It used to declare the retired REST names directly, which failed silently
 * rather than loudly (rtylr#939): every mismatched field carried a .default or
 * .catch, so zod substituted a fallback instead of erroring. channel was read as
 * `type` and lost, so TYPE_TO_CHANNEL[''] badged every campaign as Email;
 * openedCount was read as `openCount` and lost, so the open rate was
 * permanently 0%; content and segmentId were lost, so opening a campaign for
 * edit silently dropped its body and audience link.
 *
 * Mapping here rather than renaming at ~70 call sites keeps the translation in
 * one visible place. The tradeoff is that this file is now the only thing
 * standing between a contract rename and a silent regression, which is why the
 * tests assert against a wire-shaped fixture rather than a UI-shaped one.
 *
 * The backend does NOT compute openRate / clickRate / revenue — derive those
 * client-side from sentCount / openCount when a display needs them.
 */
export const campaignSchema = z
  .object({
    id: finiteNumber,
    ownerId: finiteNumber.optional(),
    name: z.string(),
    description: z.string().nullable().optional(),
    channel: z.string().default(''),
    status: z.string(),
    segmentId: finiteNumber.nullable().optional(),
    subject: z.string().nullable().optional(),
    content: z.string().nullable().optional(),
    audienceCount: finiteNumber.optional(),
    sentCount: finiteNumber.catch(0),
    openedCount: finiteNumber.catch(0),
    scheduledAt: z.string().nullable().optional(),
    sentAt: z.string().nullable().optional(),
    createdAt: z.string().optional(),
    updatedAt: z.string().optional(),
  })
  .transform(({ channel, content, openedCount, segmentId, ...rest }) => ({
    ...rest,
    type: channel,
    body: content,
    openCount: openedCount,
    audienceSegmentId: segmentId,
  }));

const legacyCampaignStatsSchema = z.object({
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

const backendCampaignStatsSchema = z.object({
  campaignId: finiteNumber,
  audienceCount: finiteNumber.catch(0),
  sentCount: finiteNumber.catch(0),
  deliveredCount: finiteNumber.catch(0),
  openedCount: finiteNumber.catch(0),
  clickedCount: finiteNumber.catch(0),
  bouncedCount: finiteNumber.catch(0),
  failedCount: finiteNumber.catch(0),
});

/** Normalize both the old UI shape and the CRM stats handler's canonical shape. */
export const campaignStatsSchema = z
  .union([legacyCampaignStatsSchema, backendCampaignStatsSchema])
  .transform((stats) => {
    if ('campaignId' in stats) {
      return {
        id: stats.campaignId,
        sent: stats.sentCount,
        delivered: stats.deliveredCount,
        opened: stats.openedCount,
        clicked: stats.clickedCount,
        unsubscribed: 0,
        bounced: stats.bouncedCount,
        revenue: 0,
        timeline: [],
      };
    }
    return stats;
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

export const CAMPAIGN_STATUSES = [
  'draft',
  'scheduled',
  'active',
  'paused',
  'completed',
  'sent',
  'cancelled',
] as const;
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
  sent: 'Sent',
  cancelled: 'Cancelled',
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
