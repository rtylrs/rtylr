import { describe, it, expect } from 'vitest';
import { classifyNps, calculateNps, npsLabel } from './npsUtils';

describe('classifyNps', () => {
  it('classifies 9-10 as promoter', () => {
    expect(classifyNps(9)).toBe('promoter');
    expect(classifyNps(10)).toBe('promoter');
  });

  it('classifies 7-8 as passive', () => {
    expect(classifyNps(7)).toBe('passive');
    expect(classifyNps(8)).toBe('passive');
  });

  it('classifies 0-6 as detractor', () => {
    expect(classifyNps(0)).toBe('detractor');
    expect(classifyNps(6)).toBe('detractor');
  });
});

describe('calculateNps', () => {
  it('returns all zeros for empty array', () => {
    const result = calculateNps([]);
    expect(result.total).toBe(0);
    expect(result.score).toBe(0);
  });

  it('calculates correct counts', () => {
    const result = calculateNps([9, 10, 8, 5]);
    expect(result.promoters).toBe(2);
    expect(result.passives).toBe(1);
    expect(result.detractors).toBe(1);
    expect(result.total).toBe(4);
  });

  it('calculates NPS score correctly', () => {
    const result = calculateNps([9, 10, 8, 5]);
    expect(result.score).toBe(25);
  });

  it('returns 100 for all promoters', () => {
    expect(calculateNps([9, 10, 9, 10]).score).toBe(100);
  });

  it('returns -100 for all detractors', () => {
    expect(calculateNps([0, 1, 3, 5]).score).toBe(-100);
  });
});

describe('npsLabel', () => {
  it('returns Excellent for score >= 70', () => {
    expect(npsLabel(70)).toBe('Excellent');
  });

  it('returns Good for score 50-69', () => {
    expect(npsLabel(50)).toBe('Good');
  });

  it('returns Fair for score 0-49', () => {
    expect(npsLabel(0)).toBe('Fair');
  });

  it('returns Poor for negative scores', () => {
    expect(npsLabel(-1)).toBe('Poor');
  });
});
