import { describe, expect, it } from 'vitest';
import {
  toLoyaltyRewardRequest,
  toLoyaltyTierRequest,
  toUpdateLoyaltyConfigRequest,
} from './useLoyalty';

describe('loyalty mutation wire mappings', () => {
  it('sends tier perks as an array instead of a quoted JSON string', () => {
    expect(
      toLoyaltyTierRequest({
        name: 'Gold',
        threshold: 1500,
        discountPct: 10,
        perks: ['Free shipping'],
        sortOrder: 2,
      })
    ).toEqual({
      name: 'Gold',
      minPoints: 1500,
      discountPct: 10,
      perks: ['Free shipping'],
      sortOrder: 2,
    });
  });

  it('prefers the edited multiplier when both discount aliases are present', () => {
    expect(
      toLoyaltyTierRequest({
        name: 'Gold',
        threshold: 1500,
        multiplier: 12,
        discountPct: 10,
      }).discountPct
    ).toBe(12);
  });

  it('fills the complete loyalty config when only one setting changes', () => {
    expect(toUpdateLoyaltyConfigRequest({ programEnabled: false })).toMatchObject({
      isEnabled: false,
      pointsPerCurrencyUnit: 1,
      currencyPerPoint: 0.01,
      bronzeMin: 0,
      silverMin: 500,
      goldMin: 1500,
      platinumMin: 5000,
      bronzeDiscountPct: 0,
      silverDiscountPct: 5,
      goldDiscountPct: 10,
      platinumDiscountPct: 15,
      pointsExpiryDays: 365,
    });
  });

  it('keeps the loyalty program enabled for a partial update by default', () => {
    expect(toUpdateLoyaltyConfigRequest({ pointsPerDollar: 2 }).isEnabled).toBe(true);
  });

  it('trims reward text fields and preserves explicit optional values', () => {
    expect(
      toLoyaltyRewardRequest({
        name: '  Free coffee  ',
        description: '  One drink  ',
        category: '  beverage  ',
        pointsCost: 100,
        stock: null,
        active: true,
      })
    ).toEqual({
      name: 'Free coffee',
      description: 'One drink',
      category: 'beverage',
      pointsCost: 100,
      stock: null,
      active: true,
    });
  });
});
