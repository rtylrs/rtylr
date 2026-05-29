import { describe, it, expect } from 'vitest';
import {
  DEFAULT_PAGE_SIZE,
  STALE_TIME_MS,
  CHART_COLORS,
  METRIC_CATEGORIES,
  METRIC_CATEGORY_LABELS,
  REVENUE_GRANULARITIES,
  REVENUE_GRANULARITY_LABELS,
  MIN_TREND_POINTS,
  TOP_LIST_LIMIT,
} from './constants';

describe('DEFAULT_PAGE_SIZE', () => {
  it('is a positive integer', () => {
    expect(DEFAULT_PAGE_SIZE).toBeGreaterThan(0);
    expect(Number.isInteger(DEFAULT_PAGE_SIZE)).toBe(true);
  });
});

describe('STALE_TIME_MS', () => {
  it('is 5 minutes in milliseconds', () => {
    expect(STALE_TIME_MS).toBe(5 * 60 * 1000);
  });
});

describe('CHART_COLORS', () => {
  it('contains at least 4 colors', () => {
    expect(CHART_COLORS.length).toBeGreaterThanOrEqual(4);
  });

  it('each color is a valid hex string', () => {
    for (const color of CHART_COLORS) {
      expect(color).toMatch(/^#[0-9a-f]{6}$/i);
    }
  });
});

describe('METRIC_CATEGORIES', () => {
  it('includes revenue, orders, customers, and staff', () => {
    expect(METRIC_CATEGORIES).toContain('revenue');
    expect(METRIC_CATEGORIES).toContain('orders');
    expect(METRIC_CATEGORIES).toContain('customers');
    expect(METRIC_CATEGORIES).toContain('staff');
  });
});

describe('METRIC_CATEGORY_LABELS', () => {
  it('has a label for every category', () => {
    for (const cat of METRIC_CATEGORIES) {
      expect(METRIC_CATEGORY_LABELS[cat]).toBeTruthy();
    }
  });
});

describe('REVENUE_GRANULARITIES', () => {
  it('contains daily, weekly, and monthly', () => {
    expect(REVENUE_GRANULARITIES).toContain('daily');
    expect(REVENUE_GRANULARITIES).toContain('weekly');
    expect(REVENUE_GRANULARITIES).toContain('monthly');
  });
});

describe('REVENUE_GRANULARITY_LABELS', () => {
  it('has a label for every granularity', () => {
    for (const g of REVENUE_GRANULARITIES) {
      expect(REVENUE_GRANULARITY_LABELS[g]).toBeTruthy();
    }
  });
});

describe('MIN_TREND_POINTS', () => {
  it('is at least 2', () => {
    expect(MIN_TREND_POINTS).toBeGreaterThanOrEqual(2);
  });
});

describe('TOP_LIST_LIMIT', () => {
  it('is a positive integer', () => {
    expect(TOP_LIST_LIMIT).toBeGreaterThan(0);
    expect(Number.isInteger(TOP_LIST_LIMIT)).toBe(true);
  });
});
