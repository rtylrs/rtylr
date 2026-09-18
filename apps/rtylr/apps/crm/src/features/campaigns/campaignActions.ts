type CampaignChannel = 'Email' | 'SMS' | 'Push';
type CampaignStatus =
  | 'Draft'
  | 'Scheduled'
  | 'Active'
  | 'Paused'
  | 'Completed'
  | 'Sent'
  | 'Cancelled';

interface CampaignActionState {
  channel: CampaignChannel;
  status: CampaignStatus;
  audienceSegmentId?: number | null;
}

/** Only channels with a dedicated editor can be safely edited in the CRM. */
export function canEditCampaign(
  campaign: Pick<CampaignActionState, 'channel' | 'status'>
): boolean {
  return (
    campaign.channel !== 'Push' && !['Completed', 'Sent', 'Cancelled'].includes(campaign.status)
  );
}

/** A campaign can only be activated after its current segment count is verified. */
export function hasVerifiedCampaignAudience(
  campaign: Pick<CampaignActionState, 'audienceSegmentId'>,
  audienceBySegment: ReadonlyMap<number, number>
): boolean {
  if (campaign.audienceSegmentId == null) return false;
  const audience = audienceBySegment.get(campaign.audienceSegmentId);
  return audience !== undefined && Number.isFinite(audience) && audience > 0;
}
