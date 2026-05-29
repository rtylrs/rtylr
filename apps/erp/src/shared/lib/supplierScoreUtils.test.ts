import { describe, it, expect } from 'vitest';
import {
  computeSupplierScore,
  scoreTier,
  rankSuppliers,
  underperformingSuppliers,
  tierColor,
  averageSupplierScore,
} from './supplierScoreUtils';
import type { Supplier } from '../services/erpApi';

const makeSupplier = (overrides: Partial<Supplier> = {}): Supplier => ({
  id: 1,
  name: 'Test Supplier',
  contactName: 'Alice',
  email: 'alice@test.com',
  phone: '555-0001',
  productCount: 5,
  categories: ['Food'],
  status: 'Active',
  onTimeDeliveryPct: 90,
  fulfillmentAccuracy: 95,
  leadTimeDays: 7,
  ...overrides,
});

describe('computeSupplierScore', () => {
  it('computes weighted score from supplier metrics', () => {
    const s = makeSupplier({ onTimeDeliveryPct: 100, fulfillmentAccuracy: 100, leadTimeDays: 0 });
    expect(computeSupplierScore(s)).toBe(100);
  });

  it('returns 0 for supplier with zero metrics and max lead time', () => {
    const s = makeSupplier({ onTimeDeliveryPct: 0, fulfillmentAccuracy: 0, leadTimeDays: 30 });
    expect(computeSupplierScore(s)).toBe(0);
  });

  it('defaults missing fields to 0', () => {
    const s = makeSupplier({
      onTimeDeliveryPct: undefined,
      fulfillmentAccuracy: undefined,
      leadTimeDays: undefined,
    });
    expect(computeSupplierScore(s)).toBe(0);
  });
});

describe('scoreTier', () => {
  it('returns excellent for score >= 85', () => {
    expect(scoreTier(85)).toBe('excellent');
    expect(scoreTier(100)).toBe('excellent');
  });

  it('returns good for score 70-84', () => {
    expect(scoreTier(70)).toBe('good');
    expect(scoreTier(84)).toBe('good');
  });

  it('returns fair for score 50-69', () => {
    expect(scoreTier(50)).toBe('fair');
    expect(scoreTier(69)).toBe('fair');
  });

  it('returns poor for score < 50', () => {
    expect(scoreTier(0)).toBe('poor');
    expect(scoreTier(49)).toBe('poor');
  });
});

describe('rankSuppliers', () => {
  it('ranks active suppliers by score descending', () => {
    const suppliers = [
      makeSupplier({ id: 1, name: 'A', onTimeDeliveryPct: 60, fulfillmentAccuracy: 60 }),
      makeSupplier({ id: 2, name: 'B', onTimeDeliveryPct: 95, fulfillmentAccuracy: 95 }),
    ];
    const ranked = rankSuppliers(suppliers);
    expect(ranked[0].supplierId).toBe(2);
    expect(ranked[1].supplierId).toBe(1);
  });

  it('excludes inactive suppliers', () => {
    const suppliers = [
      makeSupplier({ id: 1, status: 'Inactive' }),
      makeSupplier({ id: 2, status: 'Active' }),
    ];
    const ranked = rankSuppliers(suppliers);
    expect(ranked).toHaveLength(1);
    expect(ranked[0].supplierId).toBe(2);
  });
});

describe('underperformingSuppliers', () => {
  it('returns suppliers below threshold', () => {
    const suppliers = [
      makeSupplier({ id: 1, onTimeDeliveryPct: 40, fulfillmentAccuracy: 40 }),
      makeSupplier({ id: 2, onTimeDeliveryPct: 95, fulfillmentAccuracy: 95 }),
    ];
    const under = underperformingSuppliers(suppliers, 70);
    expect(under).toHaveLength(1);
    expect(under[0].supplierId).toBe(1);
  });
});

describe('tierColor', () => {
  it('returns correct color for each tier', () => {
    expect(tierColor('excellent')).toBe('green');
    expect(tierColor('good')).toBe('blue');
    expect(tierColor('fair')).toBe('yellow');
    expect(tierColor('poor')).toBe('red');
  });
});

describe('averageSupplierScore', () => {
  it('returns 0 for empty list', () => {
    expect(averageSupplierScore([])).toBe(0);
  });

  it('computes average across active suppliers', () => {
    const suppliers = [
      makeSupplier({ id: 1, onTimeDeliveryPct: 100, fulfillmentAccuracy: 100, leadTimeDays: 0 }),
      makeSupplier({ id: 2, onTimeDeliveryPct: 0, fulfillmentAccuracy: 0, leadTimeDays: 30 }),
    ];
    const avg = averageSupplierScore(suppliers);
    expect(avg).toBe(50);
  });
});
