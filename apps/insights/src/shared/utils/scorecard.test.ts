import { describe, it, expect } from 'vitest';
import {
  metricScore,
  weightedScore,
  scorecardRating,
  ratingLabel,
  underPerformingMetrics,
} from './scorecard';
import type { ScorecardMetric } from './scorecard';

function makeMet(name: string, value: number, target: number, weight = 1): ScorecardMetric {
  return { name, value, target, weight };
}

const metrics = [
  makeMet('Revenue', 90, 100, 2),
  makeMet('Retention', 60, 100, 1),
  makeMet('NPS', 80, 100, 1),
];

describe('metricScore', () => {
  it('returns percentage of target reached', () => {
    expect(metricScore(metrics[0])).toBe(90);
  });

  it('caps at 100', () => {
    expect(metricScore(makeMet('X', 120, 100))).toBe(100);
  });

  it('returns 0 when target is 0', () => {
    expect(metricScore(makeMet('X', 10, 0))).toBe(0);
  });
});

describe('weightedScore', () => {
  it('calculates weighted aggregate score', () => {
    const score = weightedScore(metrics);
    expect(score).toBeCloseTo(80, 0);
  });

  it('returns 0 for empty metrics', () => {
    expect(weightedScore([])).toBe(0);
  });
});

describe('scorecardRating', () => {
  it('returns excellent for 90+', () => {
    expect(scorecardRating(95)).toBe('excellent');
  });

  it('returns good for 70-89', () => {
    expect(scorecardRating(75)).toBe('good');
  });

  it('returns average for 50-69', () => {
    expect(scorecardRating(55)).toBe('average');
  });

  it('returns poor below 50', () => {
    expect(scorecardRating(40)).toBe('poor');
  });
});

describe('ratingLabel', () => {
  it('returns Excellent for excellent', () => {
    expect(ratingLabel('excellent')).toBe('Excellent');
  });

  it('returns Poor for poor', () => {
    expect(ratingLabel('poor')).toBe('Poor');
  });
});

describe('underPerformingMetrics', () => {
  it('returns metrics below threshold', () => {
    const under = underPerformingMetrics(metrics, 70);
    expect(under.map((m) => m.name)).toContain('Retention');
    expect(under).toHaveLength(1);
  });
});
