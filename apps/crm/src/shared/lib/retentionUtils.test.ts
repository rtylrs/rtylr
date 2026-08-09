import { describe, it, expect } from 'vitest';
import { retentionMetrics, riskLabel, clv, repeatPurchaseRate } from './retentionUtils';

describe('retentionMetrics', () => {
  it('calculates retention and churn rates', () => {
    const result = retentionMetrics(100, 80);
    expect(result.retained).toBe(80);
    expect(result.churned).toBe(20);
    expect(result.retentionRate).toBeCloseTo(0.8);
    expect(result.churnRate).toBeCloseTo(0.2);
  });

  it('returns zeros for total = 0', () => {
    const result = retentionMetrics(0, 0);
    expect(result.retentionRate).toBe(0);
    expect(result.churnRate).toBe(0);
  });

  it('caps retained at total', () => {
    const result = retentionMetrics(50, 60);
    expect(result.retained).toBe(50);
    expect(result.churned).toBe(0);
  });
});

describe('riskLabel', () => {
  it('returns healthy for <= 30 days', () => {
    expect(riskLabel(0)).toBe('healthy');
    expect(riskLabel(30)).toBe('healthy');
  });

  it('returns at-risk for 31-90 days', () => {
    expect(riskLabel(31)).toBe('at-risk');
    expect(riskLabel(90)).toBe('at-risk');
  });

  it('returns churned for > 90 days', () => {
    expect(riskLabel(91)).toBe('churned');
  });
});

describe('clv', () => {
  it('calculates customer lifetime value', () => {
    expect(clv(50, 12, 3)).toBe(1800);
  });

  it('returns 0 for zero inputs', () => {
    expect(clv(0, 12, 3)).toBe(0);
  });
});

describe('repeatPurchaseRate', () => {
  it('returns the fraction of repeat customers', () => {
    expect(repeatPurchaseRate(100, 60)).toBeCloseTo(0.6);
  });

  it('returns 0 when total is 0', () => {
    expect(repeatPurchaseRate(0, 0)).toBe(0);
  });

  it('caps at 1', () => {
    expect(repeatPurchaseRate(10, 15)).toBe(1);
  });
});
