import { describe, it, expect } from 'vitest';
import {
  filterCampaignsByStatus,
  filterCampaignsByType,
  totalSent,
  totalOpened,
  openRate,
  sortByRecency,
  isScheduledInFuture,
} from './campaignUtils';
import { campaignSchema, type Campaign } from '../schemas/campaigns';

const makeCampaign = (overrides: Record<string, unknown> = {}): Campaign =>
  campaignSchema.parse({
    id: 1,
    name: 'Test Campaign',
    type: 'email',
    status: 'active',
    sentCount: 90,
    openCount: 36,
    updatedAt: '2024-05-01',
    ...overrides,
  });

const campaigns: Campaign[] = [
  makeCampaign({ id: 1, status: 'active', type: 'email', sentCount: 90, updatedAt: '2024-05-01' }),
  makeCampaign({ id: 2, status: 'paused', type: 'sms', sentCount: 180, updatedAt: '2024-04-15' }),
  makeCampaign({ id: 3, status: 'active', type: 'email', sentCount: 140, updatedAt: '2024-05-10' }),
  makeCampaign({ id: 4, status: 'draft', type: 'push', sentCount: 0, updatedAt: '2024-03-01' }),
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

describe('filterCampaignsByType', () => {
  it('returns only email campaigns', () => {
    const result = filterCampaignsByType(campaigns, 'email');
    expect(result).toHaveLength(2);
  });

  it('returns only sms campaigns', () => {
    const result = filterCampaignsByType(campaigns, 'sms');
    expect(result).toHaveLength(1);
  });
});

describe('totalSent', () => {
  it('sums sent messages across all campaigns', () => {
    expect(totalSent(campaigns)).toBe(410);
  });

  it('returns 0 for empty array', () => {
    expect(totalSent([])).toBe(0);
  });
});

describe('totalOpened', () => {
  it('sums opened messages across all campaigns', () => {
    expect(totalOpened([makeCampaign({ openCount: 10 }), makeCampaign({ openCount: 25 })])).toBe(
      35
    );
  });
});

describe('openRate', () => {
  it('calculates rate as opened / sent', () => {
    const c = makeCampaign({ sentCount: 100, openCount: 40 });
    expect(openRate(c)).toBeCloseTo(0.4);
  });

  it('returns 0 when sent is 0', () => {
    const c = makeCampaign({ sentCount: 0, openCount: 0 });
    expect(openRate(c)).toBe(0);
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
  it('returns false when scheduledAt is not set', () => {
    const c = makeCampaign();
    expect(isScheduledInFuture(c)).toBe(false);
  });

  it('returns false for a past scheduled date', () => {
    const c = makeCampaign({ scheduledAt: '2020-01-01' });
    expect(isScheduledInFuture(c)).toBe(false);
  });

  it('returns true for a future scheduled date', () => {
    const c = makeCampaign({ scheduledAt: '2099-12-31' });
    expect(isScheduledInFuture(c)).toBe(true);
  });
});
