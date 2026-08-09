import { describe, it, expect } from 'vitest';
import { npsCategory, npsScore, npsCategoryCount, npsLabel, avgNpsScore } from './nps';
import type { NpsResponse } from './nps';

function makeResponse(score: number): NpsResponse {
  return { respondentId: 'r1', score, submittedAt: '2026-01-01' };
}

const responses = [
  makeResponse(10),
  makeResponse(9),
  makeResponse(8),
  makeResponse(7),
  makeResponse(4),
  makeResponse(2),
];

describe('npsCategory', () => {
  it('classifies 9-10 as promoter', () => {
    expect(npsCategory(9)).toBe('promoter');
    expect(npsCategory(10)).toBe('promoter');
  });

  it('classifies 7-8 as passive', () => {
    expect(npsCategory(7)).toBe('passive');
    expect(npsCategory(8)).toBe('passive');
  });

  it('classifies 0-6 as detractor', () => {
    expect(npsCategory(6)).toBe('detractor');
    expect(npsCategory(0)).toBe('detractor');
  });
});

describe('npsScore', () => {
  it('calculates NPS correctly', () => {
    const score = npsScore(responses);
    expect(score).toBe(Math.round((2 / 6) * 100 - (2 / 6) * 100));
  });

  it('returns 0 for empty responses', () => {
    expect(npsScore([])).toBe(0);
  });
});

describe('npsCategoryCount', () => {
  it('counts each category', () => {
    const counts = npsCategoryCount(responses);
    expect(counts.promoter).toBe(2);
    expect(counts.passive).toBe(2);
    expect(counts.detractor).toBe(2);
  });
});

describe('npsLabel', () => {
  it('returns Excellent for 50+', () => {
    expect(npsLabel(60)).toBe('Excellent');
  });

  it('returns Good for 30-49', () => {
    expect(npsLabel(40)).toBe('Good');
  });

  it('returns Neutral for 0-29', () => {
    expect(npsLabel(10)).toBe('Neutral');
  });

  it('returns Poor for negative', () => {
    expect(npsLabel(-10)).toBe('Poor');
  });
});

describe('avgNpsScore', () => {
  it('returns average raw score', () => {
    expect(avgNpsScore([makeResponse(10), makeResponse(6)])).toBe(8);
  });

  it('returns 0 for empty responses', () => {
    expect(avgNpsScore([])).toBe(0);
  });
});
