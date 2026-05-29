import { describe, it, expect } from 'vitest';
import {
  isCampaignStatus,
  isCampaignChannel,
  isActiveCampaign,
  campaignStatusLabel,
  campaignChannelLabel,
  CAMPAIGN_STATUS_LABELS,
  CAMPAIGN_CHANNEL_LABELS,
} from '../campaigns';

describe('isCampaignStatus', () => {
  it('returns true for valid statuses', () => {
    expect(isCampaignStatus('draft')).toBe(true);
    expect(isCampaignStatus('active')).toBe(true);
    expect(isCampaignStatus('completed')).toBe(true);
  });
  it('returns false for invalid statuses', () => {
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
    expect(isCampaignChannel('whatsapp')).toBe(false);
  });
});

describe('isActiveCampaign', () => {
  it('returns true for active campaign', () => {
    const campaign = {
      id: 1,
      name: 'Test',
      description: '',
      channel: 'email',
      status: 'active',
      audience: 0,
      sent: 0,
      openRate: '0',
      clickRate: '0',
      revenue: '0',
      updatedAt: '',
    };
    expect(isActiveCampaign(campaign)).toBe(true);
  });
  it('returns false for draft campaign', () => {
    const campaign = {
      id: 1,
      name: 'Test',
      description: '',
      channel: 'email',
      status: 'draft',
      audience: 0,
      sent: 0,
      openRate: '0',
      clickRate: '0',
      revenue: '0',
      updatedAt: '',
    };
    expect(isActiveCampaign(campaign)).toBe(false);
  });
});

describe('campaignStatusLabel', () => {
  it('returns correct labels', () => {
    expect(campaignStatusLabel('draft')).toBe('Draft');
    expect(campaignStatusLabel('active')).toBe('Active');
    expect(campaignStatusLabel('completed')).toBe('Completed');
  });
  it('covers all statuses in CAMPAIGN_STATUS_LABELS', () => {
    expect(Object.keys(CAMPAIGN_STATUS_LABELS)).toHaveLength(5);
  });
});

describe('campaignChannelLabel', () => {
  it('returns correct labels', () => {
    expect(campaignChannelLabel('email')).toBe('Email');
    expect(campaignChannelLabel('sms')).toBe('SMS');
    expect(campaignChannelLabel('push')).toBe('Push Notification');
  });
  it('covers all channels in CAMPAIGN_CHANNEL_LABELS', () => {
    expect(Object.keys(CAMPAIGN_CHANNEL_LABELS)).toHaveLength(3);
  });
});
