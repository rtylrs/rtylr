import { describe, it, expect } from 'vitest';
import {
  loyaltyConfigSchema,
  loyaltyTransactionSchema,
  loyaltyMemberSchema,
  loyaltyRewardSchema,
  loyaltyTierSchema,
  LOYALTY_TRANSACTION_TYPES,
  isEarnTransaction,
  isRedeemTransaction,
  isActiveMember,
  getTransactionIcon,
  getTransactionLabel,
} from '../loyalty';

describe('loyaltyConfigSchema', () => {
  it('parses a valid config', () => {
    const config = {
      programEnabled: true,
      pointsPerDollar: 10,
      pointsExpiry: 365,
      redemptionRate: 0.01,
    };
    const result = loyaltyConfigSchema.parse(config);
    expect(result.programEnabled).toBe(true);
    expect(result.pointsPerDollar).toBe(10);
  });
});

describe('loyaltyTransactionSchema', () => {
  it('parses an earn transaction', () => {
    const tx = { customer: 'Alice', action: 'Purchase', time: '10 min ago', type: 'earn' };
    const result = loyaltyTransactionSchema.parse(tx);
    expect(result.type).toBe('earn');
  });

  it('falls back to earn for unknown type', () => {
    const tx = { customer: 'Bob', action: 'Unknown', time: '5 min ago', type: 'bogus' };
    const result = loyaltyTransactionSchema.parse(tx);
    expect(result.type).toBe('earn');
  });

  it('parses backend upgrade transactions in the activity log', () => {
    const result = loyaltyTransactionSchema.parse({
      id: 8,
      customerId: 42,
      type: 'upgrade',
      points: 0,
      balanceBefore: 500,
      balanceAfter: 500,
      description: 'Reached Gold tier',
      createdAt: '2026-08-21T10:00:00Z',
    });
    expect(result.type).toBe('upgrade');
    expect(result.customer).toBe('42');
  });

  it('coerces backend int64 strings without losing points or balances', () => {
    const result = loyaltyTransactionSchema.parse({
      id: '9',
      customerId: '42',
      type: 'earn',
      points: '125',
      balanceBefore: '500',
      balanceAfter: '625',
      description: 'Purchase',
      createdAt: '2026-08-21T10:00:00Z',
    });
    expect(result.points).toBe(125);
    expect(result.balanceAfter).toBe(625);
  });
});

describe('backend loyalty config and tier values', () => {
  it('coerces int64 configuration and tier thresholds', () => {
    const config = loyaltyConfigSchema.parse({
      isEnabled: true,
      pointsPerCurrencyUnit: 2,
      currencyPerPoint: 0.01,
      bronzeMin: '0',
      silverMin: '500',
      goldMin: '1500',
      platinumMin: '5000',
      bronzeDiscountPct: 0,
      silverDiscountPct: 5,
      goldDiscountPct: 10,
      platinumDiscountPct: 15,
      pointsExpiryDays: '730',
    });
    const tier = loyaltyTierSchema.parse({
      id: '1',
      name: 'Gold',
      minPoints: '1500',
      discountPct: 10,
      perks: '[]',
      sortOrder: '2',
    });

    expect(config.goldMin).toBe(1500);
    expect(config.pointsExpiry).toBe(730);
    expect(tier.threshold).toBe(1500);
    expect(tier.sortOrder).toBe(2);
  });

  it('parses the mounted legacy loyalty config response', () => {
    const config = loyaltyConfigSchema.parse({
      id: '7',
      ownerId: '3',
      pointsPerDollar: 2,
      minRedemptionPoints: 100,
      pointsValue: 0.01,
      isActive: true,
      updatedAt: '2026-08-22T00:00:00Z',
    });

    expect(config.programEnabled).toBe(true);
    expect(config.pointsPerDollar).toBe(2);
    expect(config.minRedemptionPoints).toBe(100);
    expect(config.redemptionRate).toBe(0.01);
    expect(config.pointsExpiry).toBe(365);
  });
});

describe('loyaltyMemberSchema', () => {
  it('parses a valid member', () => {
    const member = {
      id: 1,
      name: 'Jane Doe',
      initials: 'JD',
      points: 1200,
      tier: 'Gold',
      visits: 24,
      joinedAt: '2023-01-01',
      lastActive: '2024-05-01',
    };
    const result = loyaltyMemberSchema.parse(member);
    expect(result.id).toBe(1);
    expect(result.tier).toBe('Gold');
  });

  it('coerces string id to number', () => {
    const member = {
      id: '99',
      name: 'John',
      initials: 'J',
      points: 500,
      tier: 'Bronze',
      visits: 5,
      joinedAt: '2024-01-01',
      lastActive: '2024-04-01',
    };
    const result = loyaltyMemberSchema.parse(member);
    expect(result.id).toBe(99);
  });

  it('treats malformed activity dates as inactive', () => {
    const member = loyaltyMemberSchema.parse({
      id: 100,
      name: 'Unknown',
      initials: 'U',
      points: 0,
      tier: 'Bronze',
      visits: 0,
      joinedAt: '2024-01-01',
      lastActive: 'not-a-date',
    });
    expect(isActiveMember(member)).toBe(false);
  });

  it('keeps members with missing activity dates renderable', () => {
    const member = loyaltyMemberSchema.parse({
      id: 101,
      name: 'No Activity',
      initials: 'NA',
      points: 0,
      tier: 'Bronze',
      visits: 0,
      joinedAt: null,
      lastActive: null,
    });
    expect(member.joinedAt).toBe('');
    expect(member.lastActive).toBe('');
    expect(isActiveMember(member)).toBe(false);
  });
});

describe('loyaltyRewardSchema', () => {
  it('parses a valid reward', () => {
    const reward = {
      id: 1,
      name: 'Free Coffee',
      description: 'Redeem for a free coffee',
      pointsCost: 200,
      category: 'beverage',
      stock: 50,
      active: true,
    };
    const result = loyaltyRewardSchema.parse(reward);
    expect(result.name).toBe('Free Coffee');
    expect(result.active).toBe(true);
  });

  it('allows null icon and handles unlimited stock (-1 or null)', () => {
    const reward = {
      id: 2,
      name: 'Discount',
      description: '10% off',
      pointsCost: 100,
      category: 'discount',
      stock: -1,
      active: true,
      icon: null,
    };
    const result = loyaltyRewardSchema.parse(reward);
    expect(result.icon).toBeNull();
    expect(result.stock).toBe(-1);
  });
});

describe('getTransactionLabel', () => {
  it('returns correct label for transaction types', () => {
    expect(getTransactionLabel('earn')).toBe('Points Earned');
    expect(getTransactionLabel('redeem')).toBe('Points Redeemed');
    expect(getTransactionLabel('award')).toBe('Points Awarded');
    expect(getTransactionLabel('adjust')).toBe('Points Adjusted');
    expect(getTransactionLabel('expire')).toBe('Points Expired');
    expect(getTransactionLabel('upgrade')).toBe('Tier Upgrade');
  });
});

describe('loyaltyTierSchema', () => {
  it('parses a valid tier with perks', () => {
    const tier = {
      id: 1,
      name: 'Gold',
      threshold: 1000,
      multiplier: 2.0,
      perks: ['Free shipping', 'Early access'],
      color: '#f59e0b',
    };
    const result = loyaltyTierSchema.parse(tier);
    expect(result.name).toBe('Gold');
    expect(result.perks).toHaveLength(2);
  });

  it('defaults perks to empty array', () => {
    const tier = {
      id: 2,
      name: 'Bronze',
      threshold: 0,
      multiplier: 1.0,
      color: '#6b7280',
    };
    const result = loyaltyTierSchema.parse(tier);
    expect(result.perks).toEqual([]);
  });
});

describe('LOYALTY_TRANSACTION_TYPES', () => {
  it('contains earn, redeem, and upgrade', () => {
    expect(LOYALTY_TRANSACTION_TYPES).toContain('earn');
    expect(LOYALTY_TRANSACTION_TYPES).toContain('redeem');
    expect(LOYALTY_TRANSACTION_TYPES).toContain('upgrade');
  });
});

describe('isEarnTransaction', () => {
  it('returns true for earn type', () => {
    const tx = loyaltyTransactionSchema.parse({
      customer: 'Alice',
      action: 'Purchase',
      time: 'now',
      type: 'earn',
    });
    expect(isEarnTransaction(tx)).toBe(true);
  });

  it('returns false for redeem type', () => {
    const tx = loyaltyTransactionSchema.parse({
      customer: 'Alice',
      action: 'Redeem',
      time: 'now',
      type: 'redeem',
    });
    expect(isEarnTransaction(tx)).toBe(false);
  });
});

describe('isRedeemTransaction', () => {
  it('returns true for redeem type', () => {
    const tx = loyaltyTransactionSchema.parse({
      customer: 'Bob',
      action: 'Redeem points',
      time: 'now',
      type: 'redeem',
    });
    expect(isRedeemTransaction(tx)).toBe(true);
  });
});

describe('getTransactionIcon', () => {
  it('returns plus for earn', () => {
    expect(getTransactionIcon('earn')).toBe('plus');
  });

  it('returns minus for redeem', () => {
    expect(getTransactionIcon('redeem')).toBe('minus');
  });

  it('returns arrow-up for upgrade', () => {
    expect(getTransactionIcon('upgrade')).toBe('arrow-up');
  });
});
