import { describe, it, expect } from 'vitest';
import { retentionRate, avgRetentionAtPeriod, cohortLabels, bestRetentionCohort } from './cohort';
import type { CohortRow } from './cohort';

function makeCohort(date: string, periods: number[]): CohortRow {
  return { cohortDate: date, periods };
}

describe('retentionRate', () => {
  it('calculates period retention as percentage of base', () => {
    const cohort = makeCohort('2026-01', [100, 80, 60]);
    expect(retentionRate(cohort, 1)).toBe(80);
    expect(retentionRate(cohort, 2)).toBe(60);
  });

  it('returns 0 when base is 0', () => {
    expect(retentionRate(makeCohort('2026-01', [0, 10]), 1)).toBe(0);
  });

  it('returns 0 for undefined period', () => {
    expect(retentionRate(makeCohort('2026-01', [100]), 5)).toBe(0);
  });
});

describe('avgRetentionAtPeriod', () => {
  it('averages retention across cohorts', () => {
    const cohorts = [makeCohort('Jan', [100, 80]), makeCohort('Feb', [100, 60])];
    expect(avgRetentionAtPeriod(cohorts, 1)).toBe(70);
  });

  it('returns 0 for empty cohorts', () => {
    expect(avgRetentionAtPeriod([], 1)).toBe(0);
  });
});

describe('cohortLabels', () => {
  it('extracts cohort dates', () => {
    const cohorts = [makeCohort('2026-01', [100]), makeCohort('2026-02', [100])];
    expect(cohortLabels(cohorts)).toEqual(['2026-01', '2026-02']);
  });
});

describe('bestRetentionCohort', () => {
  it('returns cohort with highest period-1 retention', () => {
    const cohorts = [makeCohort('Jan', [100, 60]), makeCohort('Feb', [100, 80])];
    expect(bestRetentionCohort(cohorts)?.cohortDate).toBe('Feb');
  });

  it('returns undefined for empty input', () => {
    expect(bestRetentionCohort([])).toBeUndefined();
  });
});
