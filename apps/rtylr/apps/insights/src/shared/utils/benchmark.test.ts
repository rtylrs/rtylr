import { describe, it, expect } from 'vitest';
import {
  compareToBenchmark,
  benchmarkLabel,
  topPerformers,
  underPerformers,
  rankByValue,
} from './benchmark';

describe('compareToBenchmark', () => {
  it('computes delta and pctDelta correctly', () => {
    const result = compareToBenchmark(110, 100);
    expect(result.delta).toBe(10);
    expect(result.pctDelta).toBe(10);
    expect(result.isAbove).toBe(true);
  });

  it('handles value below benchmark', () => {
    const result = compareToBenchmark(80, 100);
    expect(result.delta).toBe(-20);
    expect(result.isAbove).toBe(false);
  });

  it('returns 0 pctDelta when benchmark is 0', () => {
    const result = compareToBenchmark(50, 0);
    expect(result.pctDelta).toBe(0);
  });
});

describe('benchmarkLabel', () => {
  it('includes a sign and percentage', () => {
    const result = compareToBenchmark(110, 100);
    expect(benchmarkLabel(result)).toContain('+10%');
  });

  it('includes unit when provided', () => {
    const result = compareToBenchmark(80, 100);
    expect(benchmarkLabel(result, 'USD')).toContain('USD');
  });
});

describe('topPerformers', () => {
  it('returns items at or above threshold', () => {
    const items = [{ value: 50 }, { value: 100 }, { value: 150 }];
    expect(topPerformers(items, 100)).toHaveLength(2);
  });
});

describe('underPerformers', () => {
  it('returns items below threshold', () => {
    const items = [{ value: 50 }, { value: 100 }, { value: 150 }];
    expect(underPerformers(items, 100)).toHaveLength(1);
  });
});

describe('rankByValue', () => {
  it('sorts descending by value', () => {
    const items = [{ value: 30 }, { value: 10 }, { value: 20 }];
    const ranked = rankByValue(items);
    expect(ranked[0].value).toBe(30);
    expect(ranked[2].value).toBe(10);
  });

  it('does not mutate the original array', () => {
    const items = [{ value: 2 }, { value: 1 }];
    rankByValue(items);
    expect(items[0].value).toBe(2);
  });
});
