import { describe, it, expect } from 'vitest';
import {
  filterCampaignsByStatus,
  filterCampaignsByChannel,
  totalAudience,
  totalSent,
  deliveryRate,
  sortByRecency,
  isScheduledInFuture,
} from './campaignUtils';
import type { Campaign } from '../schemas/campaigns';

const makeCampaign = (overrides: Partial<Campaign> = {}): Campaign => ({
  id: 1,
  name: 'Test Campaign',
  description: 'A test',
  channel: 'email',
  status: 'active',
  audience: 100,
  sent: 90,
  openRate: '40%',
  clickRate: '10%',
  revenue: '$200',
  updatedAt: '2024-05-01',
  ...overrides,
});

const campaigns: Campaign[] = [
  makeCampaign({
    id: 1,
    status: 'active',
    channel: 'email',
    audience: 100,
    sent: 90,
    updatedAt: '2024-05-01',
  }),
  makeCampaign({
    id: 2,
    status: 'paused',
    channel: 'sms',
    audience: 200,
    sent: 180,
    updatedAt: '2024-04-15',
  }),
  makeCampaign({
    id: 3,
    status: 'active',
    channel: 'email',
    audience: 150,
    sent: 140,
    updatedAt: '2024-05-10',
  }),
  makeCampaign({
    id: 4,
    status: 'draft',
    channel: 'push',
    audience: 50,
    sent: 0,
    updatedAt: '2024-03-01',
  }),
];

describe('filterCampaignsByStatus', () => {
  it('returns only active campaigns', () => {
    const result = filterCampaignsByStatus(campaigns, 'active');
    expect(result).toHaveLength(2);
    expect(result.every((c) => c.status === 'active')).toBe(true);
  });

  it('returns empty array for a status with no matches', () => {
    expect(filterCampaignsByStatus(campaigns, 'completed')).toHaveLength(0);
  });
});

describe('filterCampaignsByChannel', () => {
  it('returns only email campaigns', () => {
    const result = filterCampaignsByChannel(campaigns, 'email');
    expect(result).toHaveLength(2);
  });

  it('returns only sms campaigns', () => {
    const result = filterCampaignsByChannel(campaigns, 'sms');
    expect(result).toHaveLength(1);
  });
});

describe('totalAudience', () => {
  it('sums audience across all campaigns', () => {
    expect(totalAudience(campaigns)).toBe(500);
  });

  it('returns 0 for empty array', () => {
    expect(totalAudience([])).toBe(0);
  });
});

describe('totalSent', () => {
  it('sums sent messages across all campaigns', () => {
    expect(totalSent(campaigns)).toBe(410);
  });
});

describe('deliveryRate', () => {
  it('calculates rate as sent / audience', () => {
    const c = makeCampaign({ audience: 100, sent: 90 });
    expect(deliveryRate(c)).toBeCloseTo(0.9);
  });

  it('returns 0 when audience is 0', () => {
    const c = makeCampaign({ audience: 0, sent: 0 });
    expect(deliveryRate(c)).toBe(0);
  });
});

describe('sortByRecency', () => {
  it('sorts newest updatedAt first', () => {
    const sorted = sortByRecency(campaigns);
    expect(sorted[0].updatedAt).toBe('2024-05-10');
    expect(sorted[sorted.length - 1].updatedAt).toBe('2024-03-01');
  });

  it('does not mutate the original array', () => {
    const original = [...campaigns];
    sortByRecency(campaigns);
    expect(campaigns).toEqual(original);
  });
});

describe('isScheduledInFuture', () => {
  it('returns false when scheduledFor is not set', () => {
    const c = makeCampaign();
    expect(isScheduledInFuture(c)).toBe(false);
  });

  it('returns false for a past scheduled date', () => {
    const c = makeCampaign({ scheduledFor: '2020-01-01' });
    expect(isScheduledInFuture(c)).toBe(false);
  });

  it('returns true for a future scheduled date', () => {
    const c = makeCampaign({ scheduledFor: '2099-12-31' });
    expect(isScheduledInFuture(c)).toBe(true);
  });
});
