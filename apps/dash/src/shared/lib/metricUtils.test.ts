import { describe, it, expect } from 'vitest';
import {
  trendDirection,
  clampMetric,
  shortCount,
  conversionRate,
  meetsGoal,
  goalProgress,
} from './metricUtils';

describe('trendDirection', () => {
  it('returns "up" for positive', () => expect(trendDirection(5)).toBe('up'));
  it('returns "down" for negative', () => expect(trendDirection(-3)).toBe('down'));
  it('returns "flat" for zero', () => expect(trendDirection(0)).toBe('flat'));
});

describe('clampMetric', () => {
  it('returns value within range', () => expect(clampMetric(50, 0, 100)).toBe(50));
  it('clamps to min', () => expect(clampMetric(-10, 0, 100)).toBe(0));
  it('clamps to max', () => expect(clampMetric(150, 0, 100)).toBe(100));
  it('returns min for Infinity', () => expect(clampMetric(Infinity, 0, 100)).toBe(0));
});

describe('shortCount', () => {
  it('formats 999 as "999"', () => expect(shortCount(999)).toBe('999'));
  it('formats 1000 as "1K"', () => expect(shortCount(1000)).toBe('1K'));
  it('formats 1500 as "1.5K"', () => expect(shortCount(1500)).toBe('1.5K'));
  it('formats 1M as "1M"', () => expect(shortCount(1_000_000)).toBe('1M'));
});

describe('conversionRate', () => {
  it('returns 0 for zero denominator', () => expect(conversionRate(10, 0)).toBe(0));
  it('computes rate', () => expect(conversionRate(25, 100)).toBe(25));
  it('caps at 100', () => expect(conversionRate(110, 100)).toBe(100));
});

describe('meetsGoal', () => {
  it('returns true when at goal', () => expect(meetsGoal(100, 100)).toBe(true));
  it('returns false when below goal', () => expect(meetsGoal(90, 100)).toBe(false));
  it('lower-is-better works', () => expect(meetsGoal(5, 10, false)).toBe(true));
});

describe('goalProgress', () => {
  it('returns 100 at goal', () => expect(goalProgress(100, 100)).toBe(100));
  it('returns 50 at half goal', () => expect(goalProgress(50, 100)).toBe(50));
  it('caps at 100', () => expect(goalProgress(150, 100)).toBe(100));
});
