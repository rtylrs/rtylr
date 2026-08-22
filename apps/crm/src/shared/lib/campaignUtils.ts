/**
 * Campaign utility helpers for the CRM app.
 */

import type { Campaign } from '../schemas/campaigns';
import { compareDatesDescending } from './sort';

/** Filter campaigns by status. */
export function filterCampaignsByStatus(campaigns: Campaign[], status: string): Campaign[] {
  return campaigns.filter((c) => c.status === status);
}

/** Filter campaigns by type (email / sms / push). */
export function filterCampaignsByType(campaigns: Campaign[], type: string): Campaign[] {
  return campaigns.filter((c) => c.type === type);
}

/** Sum total sent messages across all campaigns. */
export function totalSent(campaigns: Campaign[]): number {
  return campaigns.reduce((sum, c) => sum + c.sentCount, 0);
}

/** Sum total opened messages across all campaigns. */
export function totalOpened(campaigns: Campaign[]): number {
  return campaigns.reduce((sum, c) => sum + c.openCount, 0);
}

/** Return the open rate (opened / sent) for a campaign. */
export function openRate(campaign: Campaign): number {
  if (campaign.sentCount === 0) return 0;
  return campaign.openCount / campaign.sentCount;
}

/** Sort campaigns by updatedAt descending (newest first). Does not mutate. */
export function sortByRecency(campaigns: Campaign[]): Campaign[] {
  return [...campaigns].sort((a, b) => compareDatesDescending(a.updatedAt, b.updatedAt));
}

/** Return true if a campaign is scheduled for a future date. */
export function isScheduledInFuture(campaign: Campaign): boolean {
  if (!campaign.scheduledAt) return false;
  return new Date(campaign.scheduledAt).getTime() > Date.now();
}
