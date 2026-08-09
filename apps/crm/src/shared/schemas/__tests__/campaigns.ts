import { describe, it, expect } from 'vitest';
import {
  campaignSchema,
  isCampaignStatus,
  isCampaignType,
  isActiveCampaign,
  campaignStatusLabel,
  campaignTypeLabel,
  CAMPAIGN_STATUS_LABELS,
  CAMPAIGN_TYPE_LABELS,
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

describe('isCampaignType', () => {
  it('returns true for valid types', () => {
    expect(isCampaignType('email')).toBe(true);
    expect(isCampaignType('sms')).toBe(true);
    expect(isCampaignType('push')).toBe(true);
  });
  it('returns false for invalid types', () => {
    expect(isCampaignType('whatsapp')).toBe(false);
  });
});

describe('isActiveCampaign', () => {
  it('returns true for active campaign', () => {
    const campaign = campaignSchema.parse({
      id: 1,
      name: 'Test',
      type: 'email',
      status: 'active',
      sentCount: 0,
      openCount: 0,
      updatedAt: '',
    });
    expect(isActiveCampaign(campaign)).toBe(true);
  });
  it('returns false for draft campaign', () => {
    const campaign = campaignSchema.parse({
      id: 1,
      name: 'Test',
      type: 'email',
      status: 'draft',
      sentCount: 0,
      openCount: 0,
      updatedAt: '',
    });
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

describe('campaignTypeLabel', () => {
  it('returns correct labels', () => {
    expect(campaignTypeLabel('email')).toBe('Email');
    expect(campaignTypeLabel('sms')).toBe('SMS');
    expect(campaignTypeLabel('push')).toBe('Push Notification');
  });
  it('covers all types in CAMPAIGN_TYPE_LABELS', () => {
    expect(Object.keys(CAMPAIGN_TYPE_LABELS)).toHaveLength(3);
  });
});
