import { describe, expect, it } from 'vitest';
import { canEditCampaign, hasVerifiedCampaignAudience } from './campaignActions';

describe('canEditCampaign', () => {
  it('does not route Push campaigns through an email or SMS editor', () => {
    expect(canEditCampaign({ channel: 'Push', status: 'Draft' })).toBe(false);
  });

  it('allows editable email and SMS drafts', () => {
    expect(canEditCampaign({ channel: 'Email', status: 'Draft' })).toBe(true);
    expect(canEditCampaign({ channel: 'SMS', status: 'Paused' })).toBe(true);
  });

  it('keeps terminal campaigns read-only', () => {
    expect(canEditCampaign({ channel: 'Email', status: 'Completed' })).toBe(false);
    expect(canEditCampaign({ channel: 'SMS', status: 'Sent' })).toBe(false);
    expect(canEditCampaign({ channel: 'Email', status: 'Cancelled' })).toBe(false);
  });
});

describe('hasVerifiedCampaignAudience', () => {
  it('requires a known segment with a positive current count', () => {
    const audienceBySegment = new Map([
      [1, 25],
      [2, 0],
    ]);

    expect(hasVerifiedCampaignAudience({ audienceSegmentId: 1 }, audienceBySegment)).toBe(true);
    expect(hasVerifiedCampaignAudience({ audienceSegmentId: 2 }, audienceBySegment)).toBe(false);
    expect(hasVerifiedCampaignAudience({ audienceSegmentId: 3 }, audienceBySegment)).toBe(false);
    expect(hasVerifiedCampaignAudience({}, audienceBySegment)).toBe(false);
  });
});
