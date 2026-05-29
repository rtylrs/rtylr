import { describe, it, expect } from 'vitest';
import {
  campaignSchema,
  campaignStatsSchema,
  CAMPAIGN_STATUSES,
  CAMPAIGN_CHANNELS,
  isCampaignStatus,
  isCampaignChannel,
  isActiveCampaign,
  isDraftCampaign,
  isScheduledCampaign,
  isCompletedCampaign,
  campaignStatusLabel,
  campaignChannelLabel,
} from '../campaigns';

const baseCampaign = {
  id: 1,
  name: 'Summer Sale',
  description: 'Seasonal promotion',
  channel: 'email',
  status: 'active',
  audience: 500,
  sent: 480,
  openRate: '42%',
  clickRate: '12%',
  revenue: '$1,200',
  updatedAt: '2024-06-01',
};

describe('campaignSchema', () => {
  it('parses a valid campaign object', () => {
    const result = campaignSchema.parse(baseCampaign);
    expect(result.id).toBe(1);
    expect(result.name).toBe('Summer Sale');
  });

  it('coerces a string id to number', () => {
    const result = campaignSchema.parse({ ...baseCampaign, id: '42' });
    expect(result.id).toBe(42);
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

describe('isCampaignChannel', () => {
  it('validates channel strings', () => {
    expect(isCampaignChannel('email')).toBe(true);
    expect(isCampaignChannel('fax')).toBe(false);
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

describe('campaignChannelLabel', () => {
  it('returns correct labels', () => {
    expect(campaignChannelLabel('email')).toBe('Email');
    expect(campaignChannelLabel('sms')).toBe('SMS');
    expect(campaignChannelLabel('push')).toBe('Push Notification');
  });
});
