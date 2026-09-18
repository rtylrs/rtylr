import { describe, it, expect } from 'vitest';
import { resolvePeriod, periodLabel, ALL_PERIODS } from './period';

const FIXED_NOW = new Date('2026-05-15T12:00:00Z');

describe('resolvePeriod', () => {
  describe('today', () => {
    it('returns today as both from and to', () => {
      const range = resolvePeriod('today', FIXED_NOW);
      expect(range.from).toBe('2026-05-15');
      expect(range.to).toBe('2026-05-15');
    });
  });

  describe('yesterday', () => {
    it('returns the previous day as both from and to', () => {
      const range = resolvePeriod('yesterday', FIXED_NOW);
      expect(range.from).toBe('2026-05-14');
      expect(range.to).toBe('2026-05-14');
    });
  });

  describe('last_7_days', () => {
    it('spans 7 days ending today', () => {
      const range = resolvePeriod('last_7_days', FIXED_NOW);
      expect(range.from).toBe('2026-05-09');
      expect(range.to).toBe('2026-05-15');
    });
  });

  describe('last_30_days', () => {
    it('spans 30 days ending today', () => {
      const range = resolvePeriod('last_30_days', FIXED_NOW);
      expect(range.from).toBe('2026-04-16');
      expect(range.to).toBe('2026-05-15');
    });
  });

  describe('last_90_days', () => {
    it('spans 90 days ending today', () => {
      const range = resolvePeriod('last_90_days', FIXED_NOW);
      expect(range.from).toBe('2026-02-15');
      expect(range.to).toBe('2026-05-15');
    });
  });

  describe('this_week', () => {
    it('starts on Monday of the current week', () => {
      // 2026-05-15 is a Friday (dow=5), so Monday is 2026-05-11
      const range = resolvePeriod('this_week', FIXED_NOW);
      expect(range.from).toBe('2026-05-11');
      expect(range.to).toBe('2026-05-15');
    });

    it('handles Sunday (week starts previous Monday)', () => {
      const sunday = new Date('2026-05-10T12:00:00Z');
      const range = resolvePeriod('this_week', sunday);
      expect(range.from).toBe('2026-05-04');
      expect(range.to).toBe('2026-05-10');
    });
  });

  describe('this_month', () => {
    it('starts on the 1st of the current month', () => {
      const range = resolvePeriod('this_month', FIXED_NOW);
      expect(range.from).toBe('2026-05-01');
      expect(range.to).toBe('2026-05-15');
    });
  });

  describe('this_quarter', () => {
    it('Q2 starts on April 1st', () => {
      const range = resolvePeriod('this_quarter', FIXED_NOW);
      expect(range.from).toBe('2026-04-01');
      expect(range.to).toBe('2026-05-15');
    });

    it('Q1 starts on January 1st', () => {
      const q1 = new Date('2026-02-15T12:00:00Z');
      const range = resolvePeriod('this_quarter', q1);
      expect(range.from).toBe('2026-01-01');
      expect(range.to).toBe('2026-02-15');
    });
  });

  describe('this_year', () => {
    it('starts on January 1st of the current year', () => {
      const range = resolvePeriod('this_year', FIXED_NOW);
      expect(range.from).toBe('2026-01-01');
      expect(range.to).toBe('2026-05-15');
    });
  });

  describe('last_month', () => {
    it('returns the full previous month', () => {
      const range = resolvePeriod('last_month', FIXED_NOW);
      expect(range.from).toBe('2026-04-01');
      expect(range.to).toBe('2026-04-30');
    });

    it('handles January (previous month is December of last year)', () => {
      const jan = new Date('2026-01-15T12:00:00Z');
      const range = resolvePeriod('last_month', jan);
      expect(range.from).toBe('2025-12-01');
      expect(range.to).toBe('2025-12-31');
    });
  });

  describe('last_year', () => {
    it('returns the full previous year', () => {
      const range = resolvePeriod('last_year', FIXED_NOW);
      expect(range.from).toBe('2025-01-01');
      expect(range.to).toBe('2025-12-31');
    });
  });
});

describe('periodLabel', () => {
  it('returns a non-empty label for every period', () => {
    for (const period of ALL_PERIODS) {
      expect(periodLabel(period).length).toBeGreaterThan(0);
    }
  });

  it('returns "Today" for today', () => {
    expect(periodLabel('today')).toBe('Today');
  });

  it('returns "Last 30 days" for last_30_days', () => {
    expect(periodLabel('last_30_days')).toBe('Last 30 days');
  });
});

describe('ALL_PERIODS', () => {
  it('contains all 11 period ids', () => {
    expect(ALL_PERIODS).toHaveLength(11);
  });

  it('starts with today', () => {
    expect(ALL_PERIODS[0]).toBe('today');
  });

  it('ends with last_year', () => {
    expect(ALL_PERIODS[ALL_PERIODS.length - 1]).toBe('last_year');
  });
});
