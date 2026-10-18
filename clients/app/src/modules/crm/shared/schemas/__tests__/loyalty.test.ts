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

// rtylr#942: int64 travels as a string on the wire (docs/cross-app/07-plumbing.md).
// Every fixture above uses already-numeric values, which is exactly why this
// bug shipped: none of them exercise the real wire shape.
describe('loyalty schemas accept int64-as-string wire values (rtylr#942)', () => {
  it('config: parses bronze/silver/gold/platinum minimums and points expiry as wire strings', () => {
    const config = {
      isEnabled: true,
      pointsPerCurrencyUnit: 10,
      currencyPerPoint: 0.01,
      bronzeMin: '100',
      silverMin: '500',
      goldMin: '1500',
      platinumMin: '5000',
      bronzeDiscountPct: 0,
      silverDiscountPct: 5,
      goldDiscountPct: 10,
      platinumDiscountPct: 15,
      pointsExpiryDays: '365',
    };
    const result = loyaltyConfigSchema.parse(config);
    expect(result).toMatchObject({
      bronzeMin: 100,
      silverMin: 500,
      goldMin: 1500,
      platinumMin: 5000,
      pointsExpiry: 365,
    });
  });

  it('config: reads fields the wire omitted as their zero value', () => {
    // A paused program with a 0% bronze discount: is_enabled and the zero
    // discount are left out of the response entirely.
    const result = loyaltyConfigSchema.parse({
      id: '7',
      pointsPerCurrencyUnit: 2,
      currencyPerPoint: 0.01,
      silverMin: '500',
      goldMin: '1500',
      platinumMin: '5000',
      silverDiscountPct: 5,
      pointsExpiryDays: '365',
    });
    expect(result).toMatchObject({
      programEnabled: false,
      pointsPerDollar: 2,
      bronzeMin: 0,
      bronzeDiscountPct: 0,
      silverDiscountPct: 5,
      goldDiscountPct: 0,
      pointsExpiry: 365,
    });
  });

  it('transaction: parses points/balanceBefore/balanceAfter as wire strings', () => {
    const tx = {
      id: '1',
      customerId: '42',
      type: 'earn',
      points: '150',
      balanceBefore: '1000',
      balanceAfter: '1150',
      createdAt: '2026-01-01',
    };
    const result = loyaltyTransactionSchema.parse(tx);
    expect(result.points).toBe(150);
    expect(result.balanceBefore).toBe(1000);
    expect(result.balanceAfter).toBe(1150);
  });

  it('member: parses points and visits as wire strings, rather than throwing', () => {
    const member = {
      id: 1,
      name: 'Jane Doe',
      initials: 'JD',
      points: '1200',
      tier: 'Gold',
      visits: '24',
      joinedAt: '2023-01-01',
      lastActive: '2024-05-01',
    };
    const result = loyaltyMemberSchema.parse(member);
    expect(result.points).toBe(1200);
    expect(result.visits).toBe(24);
  });

  it('reward: parses pointsCost as a wire string', () => {
    const reward = {
      id: 1,
      name: 'Free Coffee',
      description: 'Redeem for a free coffee',
      pointsCost: '200',
      category: 'beverage',
      stock: 50,
      active: true,
    };
    const result = loyaltyRewardSchema.parse(reward);
    expect(result.pointsCost).toBe(200);
  });

  it('tier: parses minPoints and sortOrder as wire strings from the backend shape', () => {
    const tier = {
      id: '1',
      name: 'Gold',
      minPoints: '1000',
      discountPct: 2.0,
      perks: ['Free shipping'],
      sortOrder: '3',
    };
    const result = loyaltyTierSchema.parse(tier);
    expect(result.threshold).toBe(1000);
    expect(result.sortOrder).toBe(3);
  });
});
