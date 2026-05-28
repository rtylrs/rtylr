import { describe, it, expect } from 'vitest';
import {
  sortSuppliersByName,
  sortSuppliersByProductCount,
  sortSuppliersByOnTimeDelivery,
  filterActiveSuppliers,
  filterSuppliersByCategory,
  supplierPerformanceScore,
  supplierInitials,
} from './supplierUtils';
import type { Supplier } from '../services/erpApi';

const makeSupplier = (overrides: Partial<Supplier> = {}): Supplier => ({
  id: 1,
  name: 'Acme',
  contactName: 'Alice',
  email: 'alice@acme.com',
  phone: '555',
  productCount: 10,
  categories: ['Electronics'],
  status: 'Active',
  onTimeDeliveryPct: undefined,
  fulfillmentAccuracy: undefined,
  leadTimeDays: undefined,
  ...overrides,
});

const suppliers: Supplier[] = [
  makeSupplier({
    id: 1,
    name: 'Zebra Corp',
    productCount: 5,
    status: 'Active',
    categories: ['Tools'],
  }),
  makeSupplier({
    id: 2,
    name: 'Acme Ltd',
    productCount: 20,
    status: 'Active',
    categories: ['Electronics', 'Tools'],
  }),
  makeSupplier({
    id: 3,
    name: 'Beta Inc',
    productCount: 1,
    status: 'Inactive',
    categories: ['Electronics'],
  }),
];

describe('sortSuppliersByName', () => {
  it('sorts alphabetically ascending', () => {
    const sorted = sortSuppliersByName(suppliers);
    expect(sorted.map((s) => s.name)).toEqual(['Acme Ltd', 'Beta Inc', 'Zebra Corp']);
  });

  it('does not mutate the original array', () => {
    const original = [...suppliers];
    sortSuppliersByName(suppliers);
    expect(suppliers).toEqual(original);
  });
});

describe('sortSuppliersByProductCount', () => {
  it('sorts highest product count first', () => {
    const sorted = sortSuppliersByProductCount(suppliers);
    expect(sorted[0].productCount).toBe(20);
    expect(sorted[2].productCount).toBe(1);
  });
});

describe('sortSuppliersByOnTimeDelivery', () => {
  const withOtd: Supplier[] = [
    makeSupplier({ id: 1, name: 'A', onTimeDeliveryPct: 90 }),
    makeSupplier({ id: 2, name: 'B', onTimeDeliveryPct: undefined }),
    makeSupplier({ id: 3, name: 'C', onTimeDeliveryPct: 75 }),
  ];

  it('sorts highest OTD first', () => {
    const sorted = sortSuppliersByOnTimeDelivery(withOtd);
    expect(sorted[0].onTimeDeliveryPct).toBe(90);
    expect(sorted[1].onTimeDeliveryPct).toBe(75);
  });

  it('puts suppliers without OTD data at the end', () => {
    const sorted = sortSuppliersByOnTimeDelivery(withOtd);
    expect(sorted[2].onTimeDeliveryPct).toBeUndefined();
  });
});

describe('filterActiveSuppliers', () => {
  it('returns only active suppliers', () => {
    const active = filterActiveSuppliers(suppliers);
    expect(active).toHaveLength(2);
    expect(active.every((s) => s.status === 'Active')).toBe(true);
  });
});

describe('filterSuppliersByCategory', () => {
  it('returns suppliers that carry the category', () => {
    const result = filterSuppliersByCategory(suppliers, 'Electronics');
    expect(result).toHaveLength(2);
  });

  it('returns empty array when no match', () => {
    expect(filterSuppliersByCategory(suppliers, 'Pharma')).toHaveLength(0);
  });
});

describe('supplierPerformanceScore', () => {
  it('returns null when both metrics are missing', () => {
    expect(supplierPerformanceScore(makeSupplier())).toBeNull();
  });

  it('uses only OTD when fulfillment accuracy is missing', () => {
    const s = makeSupplier({ onTimeDeliveryPct: 80 });
    expect(supplierPerformanceScore(s)).toBe(80);
  });

  it('uses only fulfillment accuracy when OTD is missing', () => {
    const s = makeSupplier({ fulfillmentAccuracy: 70 });
    expect(supplierPerformanceScore(s)).toBe(70);
  });

  it('computes weighted score: OTD 60% + FA 40%', () => {
    const s = makeSupplier({ onTimeDeliveryPct: 100, fulfillmentAccuracy: 80 });
    expect(supplierPerformanceScore(s)).toBe(92);
  });

  it('returns 0 for a supplier with 0 on both metrics', () => {
    const s = makeSupplier({ onTimeDeliveryPct: 0, fulfillmentAccuracy: 0 });
    expect(supplierPerformanceScore(s)).toBe(0);
  });
});

describe('supplierInitials', () => {
  it('returns first two initials uppercased', () => {
    expect(supplierInitials('Acme Ltd')).toBe('AL');
  });

  it('returns a single letter for a one-word name', () => {
    expect(supplierInitials('Acme')).toBe('A');
  });

  it('truncates to two characters for multi-word names', () => {
    expect(supplierInitials('One Two Three')).toBe('OT');
  });
});
