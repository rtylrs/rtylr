import { describe, it, expect } from 'vitest';
import {
  DEFAULT_PAGE_SIZE,
  PAGE_SIZE_OPTIONS,
  DEFAULT_LOW_STOCK_THRESHOLD,
  DEFAULT_REORDER_SAFETY_FACTOR,
  DEFAULT_LEAD_TIME_DAYS,
  PRICE_LIST_TIERS,
  PO_STATUSES,
  TRANSFER_STATUSES,
  STOCK_TAKE_STATUSES,
  SUPPLIER_STATUSES,
  VALUATION_METHODS,
  ACTIVITY_TYPES,
  CONFIDENCE_LEVELS,
  STALE_TIME_PRODUCTS,
  STALE_TIME_DASHBOARD,
} from './constants';

describe('pagination constants', () => {
  it('DEFAULT_PAGE_SIZE is a positive integer', () => {
    expect(DEFAULT_PAGE_SIZE).toBeGreaterThan(0);
    expect(Number.isInteger(DEFAULT_PAGE_SIZE)).toBe(true);
  });

  it('PAGE_SIZE_OPTIONS is sorted ascending', () => {
    const sorted = [...PAGE_SIZE_OPTIONS].sort((a, b) => a - b);
    expect([...PAGE_SIZE_OPTIONS]).toEqual(sorted);
  });

  it('PAGE_SIZE_OPTIONS includes DEFAULT_PAGE_SIZE', () => {
    expect(PAGE_SIZE_OPTIONS).toContain(DEFAULT_PAGE_SIZE);
  });
});

describe('stock constants', () => {
  it('DEFAULT_LOW_STOCK_THRESHOLD is positive', () => {
    expect(DEFAULT_LOW_STOCK_THRESHOLD).toBeGreaterThan(0);
  });

  it('DEFAULT_REORDER_SAFETY_FACTOR is between 0 and 1', () => {
    expect(DEFAULT_REORDER_SAFETY_FACTOR).toBeGreaterThanOrEqual(0);
    expect(DEFAULT_REORDER_SAFETY_FACTOR).toBeLessThanOrEqual(1);
  });

  it('DEFAULT_LEAD_TIME_DAYS is a positive integer', () => {
    expect(DEFAULT_LEAD_TIME_DAYS).toBeGreaterThan(0);
    expect(Number.isInteger(DEFAULT_LEAD_TIME_DAYS)).toBe(true);
  });
});

describe('status tuple constants', () => {
  it('PRICE_LIST_TIERS includes expected values', () => {
    expect(PRICE_LIST_TIERS).toContain('Retail');
    expect(PRICE_LIST_TIERS).toContain('Wholesale');
  });

  it('PO_STATUSES includes Draft and Approved', () => {
    expect(PO_STATUSES).toContain('Draft');
    expect(PO_STATUSES).toContain('Approved');
  });

  it('TRANSFER_STATUSES includes In Transit and Received', () => {
    expect(TRANSFER_STATUSES).toContain('In Transit');
    expect(TRANSFER_STATUSES).toContain('Received');
  });

  it('STOCK_TAKE_STATUSES includes all three statuses', () => {
    expect(STOCK_TAKE_STATUSES).toHaveLength(3);
  });

  it('SUPPLIER_STATUSES includes Active and Inactive', () => {
    expect(SUPPLIER_STATUSES).toContain('Active');
    expect(SUPPLIER_STATUSES).toContain('Inactive');
  });

  it('VALUATION_METHODS includes FIFO, LIFO, WAC', () => {
    expect(VALUATION_METHODS).toContain('FIFO');
    expect(VALUATION_METHODS).toContain('LIFO');
    expect(VALUATION_METHODS).toContain('WAC');
  });

  it('ACTIVITY_TYPES includes product and order', () => {
    expect(ACTIVITY_TYPES).toContain('product');
    expect(ACTIVITY_TYPES).toContain('order');
  });

  it('CONFIDENCE_LEVELS includes high, medium, low', () => {
    expect(CONFIDENCE_LEVELS).toContain('high');
    expect(CONFIDENCE_LEVELS).toContain('medium');
    expect(CONFIDENCE_LEVELS).toContain('low');
  });
});

describe('stale time constants', () => {
  it('STALE_TIME_PRODUCTS is a positive number', () => {
    expect(STALE_TIME_PRODUCTS).toBeGreaterThan(0);
  });

  it('STALE_TIME_DASHBOARD is less than STALE_TIME_PRODUCTS', () => {
    expect(STALE_TIME_DASHBOARD).toBeLessThan(STALE_TIME_PRODUCTS);
  });
});
