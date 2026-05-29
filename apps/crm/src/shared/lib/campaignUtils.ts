/**
 * Campaign utility helpers for the CRM app.
 */

import type { Campaign } from '../schemas/campaigns';

/** Filter campaigns by status. */
export function filterCampaignsByStatus(campaigns: Campaign[], status: string): Campaign[] {
  return campaigns.filter((c) => c.status === status);
}

/** Filter campaigns by channel. */
export function filterCampaignsByChannel(campaigns: Campaign[], channel: string): Campaign[] {
  return campaigns.filter((c) => c.channel === channel);
}

/** Sum total audience across all campaigns. */
export function totalAudience(campaigns: Campaign[]): number {
  return campaigns.reduce((sum, c) => sum + c.audience, 0);
}

/** Sum total sent messages across all campaigns. */
export function totalSent(campaigns: Campaign[]): number {
  return campaigns.reduce((sum, c) => sum + c.sent, 0);
}

/** Return the delivery rate (sent / audience) for a campaign. */
export function deliveryRate(campaign: Campaign): number {
  if (campaign.audience === 0) return 0;
  return campaign.sent / campaign.audience;
}

/** Sort campaigns by updatedAt descending (newest first). Does not mutate. */
export function sortByRecency(campaigns: Campaign[]): Campaign[] {
  return [...campaigns].sort(
    (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
  );
}

/** Return true if a campaign is scheduled for a future date. */
export function isScheduledInFuture(campaign: Campaign): boolean {
  if (!campaign.scheduledFor) return false;
  return new Date(campaign.scheduledFor).getTime() > Date.now();
}
