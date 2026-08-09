import { describe, it, expect } from 'vitest';
import {
  criteriaScore,
  weightedCandidateScore,
  passesThreshold,
  weakCriteria,
  topCriteria,
  scorecardSummary,
} from './scorecard';
import type { ScorecardCriteria } from './scorecard';

function makeCriteria(
  name: string,
  score: number,
  maxScore: number,
  weight = 1
): ScorecardCriteria {
  return { id: name, name, weight, score, maxScore };
}

const criteria = [
  makeCriteria('Communication', 8, 10, 2),
  makeCriteria('Technical', 7, 10, 3),
  makeCriteria('Culture', 9, 10, 1),
];

describe('criteriaScore', () => {
  it('returns percentage score', () => {
    expect(criteriaScore(makeCriteria('X', 8, 10))).toBe(80);
  });

  it('returns 0 when maxScore is 0', () => {
    expect(criteriaScore(makeCriteria('X', 5, 0))).toBe(0);
  });
});

describe('weightedCandidateScore', () => {
  it('returns weighted total score', () => {
    const score = weightedCandidateScore(criteria);
    expect(score).toBeGreaterThan(0);
    expect(score).toBeLessThanOrEqual(100);
  });

  it('returns 0 for empty criteria', () => {
    expect(weightedCandidateScore([])).toBe(0);
  });
});

describe('passesThreshold', () => {
  it('returns true when score meets threshold', () => {
    expect(passesThreshold(criteria, 70)).toBe(true);
  });

  it('returns false when score is below threshold', () => {
    expect(passesThreshold(criteria, 99)).toBe(false);
  });
});

describe('weakCriteria', () => {
  it('returns criteria below threshold', () => {
    const weak = weakCriteria([makeCriteria('A', 5, 10), makeCriteria('B', 9, 10)], 60);
    expect(weak).toHaveLength(1);
    expect(weak[0].name).toBe('A');
  });
});

describe('topCriteria', () => {
  it('returns criteria at or above threshold', () => {
    const top = topCriteria([makeCriteria('A', 8, 10), makeCriteria('B', 7, 10)], 80);
    expect(top).toHaveLength(1);
    expect(top[0].name).toBe('A');
  });
});

describe('scorecardSummary', () => {
  it('returns Recommended when score passes threshold', () => {
    expect(scorecardSummary(criteria, 70)).toContain('Recommended');
  });

  it('returns Not Recommended when score fails threshold', () => {
    expect(scorecardSummary(criteria, 99)).toContain('Not Recommended');
  });
});
