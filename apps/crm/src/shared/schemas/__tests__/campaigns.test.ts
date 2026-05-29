import { describe, it, expect } from 'vitest';
import {
  campaignSchema,
  campaignStatsSchema,
  CAMPAIGN_STATUSES,
  CAMPAIGN_CHANNELS,
  isCampaignStatus,
  isCampaignChannel,
  isActiveCampaign,
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

  it('allows nullable optional fields', () => {
    const result = campaignSchema.parse({ ...baseCampaign, subject: null, body: null });
    expect(result.subject).toBeNull();
    expect(result.body).toBeNull();
  });

  it('allows scheduledFor to be omitted', () => {
    const result = campaignSchema.parse(baseCampaign);
    expect(result.scheduledFor).toBeUndefined();
  });
});

describe('campaignStatsSchema', () => {
  it('parses valid stats with a timeline', () => {
    const stats = {
      id: 1,
      sent: 100,
      delivered: 95,
      opened: 40,
      clicked: 10,
      unsubscribed: 2,
      bounced: 5,
      revenue: 500,
      timeline: [{ day: '2024-06-01', opens: 20, clicks: 5 }],
    };
    const result = campaignStatsSchema.parse(stats);
    expect(result.timeline).toHaveLength(1);
  });

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
    expect(CAMPAIGN_STATUSES).toContain('scheduled');
    expect(CAMPAIGN_STATUSES).toContain('active');
    expect(CAMPAIGN_STATUSES).toContain('paused');
    expect(CAMPAIGN_STATUSES).toContain('completed');
  });
});

describe('CAMPAIGN_CHANNELS', () => {
  it('contains email, sms, and push', () => {
    expect(CAMPAIGN_CHANNELS).toContain('email');
    expect(CAMPAIGN_CHANNELS).toContain('sms');
    expect(CAMPAIGN_CHANNELS).toContain('push');
  });
});

describe('isCampaignStatus', () => {
  it('returns true for valid statuses', () => {
    expect(isCampaignStatus('active')).toBe(true);
    expect(isCampaignStatus('draft')).toBe(true);
  });

  it('returns false for unknown strings', () => {
    expect(isCampaignStatus('unknown')).toBe(false);
    expect(isCampaignStatus('')).toBe(false);
  });
});

describe('isCampaignChannel', () => {
  it('returns true for valid channels', () => {
    expect(isCampaignChannel('email')).toBe(true);
    expect(isCampaignChannel('sms')).toBe(true);
    expect(isCampaignChannel('push')).toBe(true);
  });

  it('returns false for invalid channels', () => {
    expect(isCampaignChannel('fax')).toBe(false);
  });
});

describe('isActiveCampaign', () => {
  it('returns true when status is active', () => {
    const campaign = campaignSchema.parse({ ...baseCampaign, status: 'active' });
    expect(isActiveCampaign(campaign)).toBe(true);
  });

  it('returns false when status is not active', () => {
    const campaign = campaignSchema.parse({ ...baseCampaign, status: 'paused' });
    expect(isActiveCampaign(campaign)).toBe(false);
  });
});
