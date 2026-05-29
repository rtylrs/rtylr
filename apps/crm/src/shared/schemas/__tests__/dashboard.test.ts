import { describe, it, expect } from 'vitest';
import { topCustomerSchema, recentCampaignSchema, dashboardStatsSchema } from '../dashboard';

const baseStats = {
  totalCustomers: 1200,
  newThisMonth: 45,
  activeLoyaltyMembers: 380,
  campaignReach: 950,
  totalCustomersChange: '+3%',
  newThisMonthChange: '+12%',
  activeLoyaltyMembersChange: '+5%',
  campaignReachChange: '-2%',
};

describe('topCustomerSchema', () => {
  it('parses a valid top customer', () => {
    const result = topCustomerSchema.parse({
      name: 'Alice',
      spend: '$1,200',
      visits: 24,
      tier: 'Gold',
    });
    expect(result.name).toBe('Alice');
    expect(result.tier).toBe('Gold');
  });
});

describe('recentCampaignSchema', () => {
  it('parses a valid recent campaign', () => {
    const result = recentCampaignSchema.parse({
      name: 'Summer Promo',
      channel: 'email',
      sent: 500,
      openRate: '38%',
      revenue: '$800',
    });
    expect(result.name).toBe('Summer Promo');
    expect(result.sent).toBe(500);
  });
});

describe('dashboardStatsSchema', () => {
  it('parses a minimal stats object', () => {
    const result = dashboardStatsSchema.parse(baseStats);
    expect(result.totalCustomers).toBe(1200);
    expect(result.newThisMonth).toBe(45);
  });

  it('defaults topCustomers to empty array', () => {
    const result = dashboardStatsSchema.parse(baseStats);
    expect(result.topCustomers).toEqual([]);
  });

  it('defaults recentCampaigns to empty array', () => {
    const result = dashboardStatsSchema.parse(baseStats);
    expect(result.recentCampaigns).toEqual([]);
  });

  it('parses with topCustomers and recentCampaigns', () => {
    const result = dashboardStatsSchema.parse({
      ...baseStats,
      topCustomers: [{ name: 'Bob', spend: '$500', visits: 10, tier: 'Silver' }],
      recentCampaigns: [
        { name: 'Loyalty Push', channel: 'sms', sent: 200, openRate: '45%', revenue: '$350' },
      ],
    });
    expect(result.topCustomers).toHaveLength(1);
    expect(result.recentCampaigns).toHaveLength(1);
  });

  it('includes all change fields', () => {
    const result = dashboardStatsSchema.parse(baseStats);
    expect(result.totalCustomersChange).toBe('+3%');
    expect(result.campaignReachChange).toBe('-2%');
  });
});
