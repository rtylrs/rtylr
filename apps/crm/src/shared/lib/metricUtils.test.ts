import { describe, it, expect } from 'vitest';
import {
  formatMetric,
  absoluteDiff,
  pctChange,
  formatPctChange,
  clamp,
  average,
  median,
  round,
} from './metricUtils';

describe('formatMetric', () => {
  it('formats millions', () => expect(formatMetric(2_500_000)).toBe('2.5M'));
  it('formats thousands', () => expect(formatMetric(1500)).toBe('1.5K'));
  it('formats small numbers', () => expect(formatMetric(42)).toBe('42'));
});

describe('absoluteDiff', () => {
  it('returns positive diff', () => expect(absoluteDiff(10, 3)).toBe(7));
  it('handles negative inputs', () => expect(absoluteDiff(3, 10)).toBe(7));
});

describe('pctChange', () => {
  it('calculates percentage increase', () => expect(pctChange(100, 150)).toBe(50));
  it('calculates percentage decrease', () => expect(pctChange(200, 100)).toBe(-50));
  it('returns 0 when old value is 0', () => expect(pctChange(0, 100)).toBe(0));
});

describe('formatPctChange', () => {
  it('formats positive change with plus sign', () =>
    expect(formatPctChange(100, 150)).toBe('+50.0%'));
  it('formats negative change', () => expect(formatPctChange(200, 100)).toBe('-50.0%'));
});

describe('clamp', () => {
  it('clamps to min', () => expect(clamp(-5, 0, 10)).toBe(0));
  it('clamps to max', () => expect(clamp(15, 0, 10)).toBe(10));
  it('returns value when in range', () => expect(clamp(5, 0, 10)).toBe(5));
});

describe('average', () => {
  it('calculates average', () => expect(average([1, 2, 3, 4, 5])).toBe(3));
  it('returns 0 for empty array', () => expect(average([])).toBe(0));
});

describe('median', () => {
  it('calculates median for odd count', () => expect(median([1, 3, 5])).toBe(3));
  it('calculates median for even count', () => expect(median([1, 2, 3, 4])).toBe(2.5));
  it('returns 0 for empty array', () => expect(median([])).toBe(0));
});

describe('round', () => {
  it('rounds to 2 decimal places by default', () => expect(round(1.2345)).toBe(1.23));
  it('rounds to specified places', () => expect(round(1.2345, 3)).toBe(1.235));
});
