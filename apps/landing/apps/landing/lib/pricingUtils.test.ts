import { describe, it, expect } from 'vitest';
import {
  getPlanPrice,
  getPlanPriceNote,
  isFeaturedPlan,
  getFeaturedPlan,
  annualSavingsPct,
  filterPlansByAction,
  isFreePlan,
} from './pricingUtils';
import type { Plan } from '../components/pricing/pricingData';

function makePlan(overrides: Partial<Plan> = {}): Plan {
  return {
    name: 'Test',
    price: { monthly: '$10', annual: '$8' },
    priceSuffix: '/seat/month',
    priceNote: { monthly: 'monthly note', annual: 'annual note' },
    description: 'A test plan',
    features: ['Feature A'],
    ctaText: 'Get started',
    ctaAction: 'register',
    ...overrides,
  };
}

describe('getPlanPrice', () => {
  it('returns monthly price for monthly cycle', () => {
    expect(getPlanPrice(makePlan(), 'monthly')).toBe('$10');
  });

  it('returns annual price for annual cycle', () => {
    expect(getPlanPrice(makePlan(), 'annual')).toBe('$8');
  });
});

describe('getPlanPriceNote', () => {
  it('returns the monthly note for monthly cycle', () => {
    expect(getPlanPriceNote(makePlan(), 'monthly')).toBe('monthly note');
  });

  it('returns the annual note for annual cycle', () => {
    expect(getPlanPriceNote(makePlan(), 'annual')).toBe('annual note');
  });
});

describe('isFeaturedPlan', () => {
  it('returns true when featured is true', () => {
    expect(isFeaturedPlan(makePlan({ featured: true }))).toBe(true);
  });

  it('returns false when featured is undefined', () => {
    expect(isFeaturedPlan(makePlan())).toBe(false);
  });

  it('returns false when featured is false', () => {
    expect(isFeaturedPlan(makePlan({ featured: false }))).toBe(false);
  });
});

describe('getFeaturedPlan', () => {
  it('returns the featured plan', () => {
    const plans = [makePlan({ name: 'Free' }), makePlan({ name: 'Pro', featured: true })];
    expect(getFeaturedPlan(plans)?.name).toBe('Pro');
  });

  it('returns undefined when no plan is featured', () => {
    expect(getFeaturedPlan([makePlan(), makePlan()])).toBeUndefined();
  });
});

describe('annualSavingsPct', () => {
  it('returns the percentage saving', () => {
    const plan = makePlan({ price: { monthly: '$10', annual: '$8' } });
    expect(annualSavingsPct(plan)).toBe(20);
  });

  it('returns null for a free plan with $0 monthly price', () => {
    const plan = makePlan({ price: { monthly: '$0', annual: '$0' } });
    expect(annualSavingsPct(plan)).toBeNull();
  });

  it('rounds to the nearest percent', () => {
    const plan = makePlan({ price: { monthly: '$6.99', annual: '$4.99' } });
    const savings = annualSavingsPct(plan);
    expect(savings).toBeGreaterThan(0);
    expect(typeof savings).toBe('number');
  });
});

describe('filterPlansByAction', () => {
  it('returns plans matching the register action', () => {
    const plans = [
      makePlan({ ctaAction: 'register' }),
      makePlan({ ctaAction: 'contact' }),
      makePlan({ ctaAction: 'register' }),
    ];
    expect(filterPlansByAction(plans, 'register')).toHaveLength(2);
  });

  it('returns empty array when no match', () => {
    const plans = [makePlan({ ctaAction: 'register' })];
    expect(filterPlansByAction(plans, 'contact')).toHaveLength(0);
  });
});

describe('isFreePlan', () => {
  it('returns true for a $0 plan', () => {
    const plan = makePlan({ price: { monthly: '$0', annual: '$0' } });
    expect(isFreePlan(plan, 'monthly')).toBe(true);
  });

  it('returns false for a paid plan', () => {
    expect(isFreePlan(makePlan(), 'monthly')).toBe(false);
  });
});
