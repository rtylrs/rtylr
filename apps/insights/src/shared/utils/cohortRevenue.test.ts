import { describe, it, expect } from 'vitest';
import {
  revenuePerUser,
  ltv,
  bestRevenuePeriod,
  avgCohortRevenue,
  highestLtvCohort,
} from './cohortRevenue';
import type { CohortRevenuePeriod } from './cohortRevenue';

function makeCohort(
  cohortDate: string,
  periods: number[],
  initialSize: number
): CohortRevenuePeriod {
  return { cohortDate, periods, initialSize };
}

const cohort = makeCohort('2026-01', [1000, 800, 600, 400], 100);

describe('revenuePerUser', () => {
  it('returns revenue per user for period', () => {
    expect(revenuePerUser(cohort, 0)).toBe(10);
    expect(revenuePerUser(cohort, 1)).toBe(8);
  });

  it('returns 0 when initial size is 0', () => {
    expect(revenuePerUser(makeCohort('X', [100], 0), 0)).toBe(0);
  });

  it('returns 0 for undefined period', () => {
    expect(revenuePerUser(cohort, 99)).toBe(0);
  });
});

describe('ltv', () => {
  it('returns cumulative LTV up to period', () => {
    expect(ltv(cohort, 1)).toBe(18);
    expect(ltv(cohort, 3)).toBe(28);
  });

  it('returns 0 when initial size is 0', () => {
    expect(ltv(makeCohort('X', [100], 0), 0)).toBe(0);
  });
});

describe('bestRevenuePeriod', () => {
  it('returns index of highest revenue period', () => {
    expect(bestRevenuePeriod(cohort)).toBe(0);
  });

  it('returns -1 for empty periods', () => {
    expect(bestRevenuePeriod(makeCohort('X', [], 100))).toBe(-1);
  });
});

describe('avgCohortRevenue', () => {
  it('returns average revenue across cohorts at a period', () => {
    const cohorts = [makeCohort('Jan', [1000, 800], 100), makeCohort('Feb', [1200, 900], 100)];
    expect(avgCohortRevenue(cohorts, 0)).toBe(1100);
  });

  it('returns 0 for empty cohorts', () => {
    expect(avgCohortRevenue([], 0)).toBe(0);
  });
});

describe('highestLtvCohort', () => {
  it('returns cohort with highest LTV', () => {
    const cohorts = [makeCohort('Jan', [1000, 800], 100), makeCohort('Feb', [500, 400], 100)];
    expect(highestLtvCohort(cohorts, 1)?.cohortDate).toBe('Jan');
  });

  it('returns undefined for empty cohorts', () => {
    expect(highestLtvCohort([], 0)).toBeUndefined();
  });
});
