import { describe, it, expect } from 'vitest';
import {
  campaignSchema,
  campaignStatsSchema,
  calculateOpenRate,
  CAMPAIGN_STATUSES,
  CAMPAIGN_TYPES,
  isCampaignStatus,
  isCampaignType,
  isActiveCampaign,
  isDraftCampaign,
  isScheduledCampaign,
  isCompletedCampaign,
  campaignStatusLabel,
  campaignTypeLabel,
} from '../campaigns';

// Contract field names, because that is what CampaignServiceClient returns.
// The previous fixture used the retired REST names, so it agreed with the schema
// and disagreed with the server, which is why the mismatch in rtylr#939 survived
// a green suite.
const baseCampaign = {
  id: 1,
  name: 'Summer Sale',
  channel: 'email',
  status: 'active',
  segmentId: 3,
  subject: 'Big summer savings',
  content: 'Shop now',
  sentCount: 480,
  openedCount: 200,
  updatedAt: '2024-06-01',
};

describe('campaignSchema', () => {
  it('parses a valid campaign object', () => {
    const result = campaignSchema.parse(baseCampaign);
    expect(result.id).toBe(1);
    expect(result.name).toBe('Summer Sale');
    expect(result.type).toBe('email');
    expect(result.sentCount).toBe(480);
    expect(result.openCount).toBe(200);
  });

  it('coerces a string id to number', () => {
    const result = campaignSchema.parse({ ...baseCampaign, id: '42' });
    expect(result.id).toBe(42);
  });
});

describe('calculateOpenRate', () => {
  it('derives the open rate as a percentage', () => {
    const result = campaignSchema.parse({ ...baseCampaign, sentCount: 400, openedCount: 100 });
    expect(calculateOpenRate(result)).toBe(25);
  });

  it('returns 0 when nothing was sent', () => {
    const result = campaignSchema.parse({ ...baseCampaign, sentCount: 0, openedCount: 0 });
    expect(calculateOpenRate(result)).toBe(0);
  });
});

describe('campaignStatsSchema', () => {
  it('defaults timeline to empty array when missing', () => {
    const stats = {
      id: 1,
      sent: 100,
      delivered: 95,
      opened: 40,
      clicked: 10,
      unsubscribed: 2,
      bounced: 5,
      revenue: 500,
    };
    const result = campaignStatsSchema.parse(stats);
    expect(result.timeline).toEqual([]);
  });
});

describe('CAMPAIGN_STATUSES', () => {
  it('contains all expected statuses', () => {
    expect(CAMPAIGN_STATUSES).toContain('draft');
    expect(CAMPAIGN_STATUSES).toContain('active');
    expect(CAMPAIGN_STATUSES).toContain('completed');
  });
});

describe('isCampaignStatus', () => {
  it('validates status strings', () => {
    expect(isCampaignStatus('active')).toBe(true);
    expect(isCampaignStatus('unknown')).toBe(false);
  });
});

describe('CAMPAIGN_TYPES', () => {
  it('contains the supported campaign types', () => {
    expect(CAMPAIGN_TYPES).toContain('email');
    expect(CAMPAIGN_TYPES).toContain('sms');
    expect(CAMPAIGN_TYPES).toContain('push');
  });
});

describe('isCampaignType', () => {
  it('validates type strings', () => {
    expect(isCampaignType('email')).toBe(true);
    expect(isCampaignType('fax')).toBe(false);
  });
});

describe('isActiveCampaign', () => {
  it('returns true for active status', () => {
    expect(isActiveCampaign(campaignSchema.parse({ ...baseCampaign, status: 'active' }))).toBe(
      true
    );
  });
  it('returns false for non-active', () => {
    expect(isActiveCampaign(campaignSchema.parse({ ...baseCampaign, status: 'paused' }))).toBe(
      false
    );
  });
});

describe('isDraftCampaign', () => {
  it('returns true for draft status', () => {
    expect(isDraftCampaign(campaignSchema.parse({ ...baseCampaign, status: 'draft' }))).toBe(true);
  });
});

describe('isScheduledCampaign', () => {
  it('returns true for scheduled status', () => {
    expect(
      isScheduledCampaign(campaignSchema.parse({ ...baseCampaign, status: 'scheduled' }))
    ).toBe(true);
  });
});

describe('isCompletedCampaign', () => {
  it('returns true for completed status', () => {
    expect(
      isCompletedCampaign(campaignSchema.parse({ ...baseCampaign, status: 'completed' }))
    ).toBe(true);
  });
});

describe('campaignStatusLabel', () => {
  it('returns correct labels', () => {
    expect(campaignStatusLabel('draft')).toBe('Draft');
    expect(campaignStatusLabel('active')).toBe('Active');
    expect(campaignStatusLabel('completed')).toBe('Completed');
  });
});

describe('campaignTypeLabel', () => {
  it('returns correct labels', () => {
    expect(campaignTypeLabel('email')).toBe('Email');
    expect(campaignTypeLabel('sms')).toBe('SMS');
    expect(campaignTypeLabel('push')).toBe('Push Notification');
  });
});
