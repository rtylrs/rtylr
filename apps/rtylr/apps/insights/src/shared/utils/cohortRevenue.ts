// Cohort revenue analysis helpers for the insights analytics feature.

export interface CohortRevenuePeriod {
  cohortDate: string;
  periods: number[];
  initialSize: number;
}

/**
 * revenuePerUser returns the revenue per user for a specific period.
 */
export function revenuePerUser(cohort: CohortRevenuePeriod, period: number): number {
  if (cohort.initialSize === 0) return 0;
  const periodRevenue = cohort.periods[period];
  if (periodRevenue === undefined) return 0;
  return Math.round((periodRevenue / cohort.initialSize) * 100) / 100;
}

/**
 * ltv returns the cumulative lifetime value per user up to the given period.
 */
export function ltv(cohort: CohortRevenuePeriod, upToPeriod: number): number {
  if (cohort.initialSize === 0) return 0;
  const total = cohort.periods.slice(0, upToPeriod + 1).reduce((sum, v) => sum + v, 0);
  return Math.round((total / cohort.initialSize) * 100) / 100;
}

/**
 * bestRevenuePeriod returns the period index with the highest total revenue.
 */
export function bestRevenuePeriod(cohort: CohortRevenuePeriod): number {
  if (cohort.periods.length === 0) return -1;
  return cohort.periods.reduce((best, v, i) => (v > cohort.periods[best] ? i : best), 0);
}

/**
 * avgCohortRevenue returns the average period revenue across all cohorts at a given period.
 */
export function avgCohortRevenue(cohorts: CohortRevenuePeriod[], period: number): number {
  const valid = cohorts.filter((c) => c.periods[period] !== undefined);
  if (valid.length === 0) return 0;
  const sum = valid.reduce((acc, c) => acc + (c.periods[period] ?? 0), 0);
  return Math.round((sum / valid.length) * 100) / 100;
}

/**
 * highestLtvCohort returns the cohort with the highest cumulative LTV at a given period.
 */
export function highestLtvCohort(
  cohorts: CohortRevenuePeriod[],
  period: number
): CohortRevenuePeriod | undefined {
  if (cohorts.length === 0) return undefined;
  return cohorts.reduce((best, c) => (ltv(c, period) > ltv(best, period) ? c : best));
}
