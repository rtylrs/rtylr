// Cohort analysis helpers for the insights customer feature.

export interface CohortRow {
  cohortDate: string;
  periods: number[];
}

/**
 * retentionRate returns the retention percentage for a specific period.
 * Returns 0 when the base count is 0.
 */
export function retentionRate(cohort: CohortRow, period: number): number {
  const base = cohort.periods[0];
  const current = cohort.periods[period];
  if (base === 0 || current === undefined) return 0;
  return Math.round((current / base) * 100 * 10) / 10;
}

/**
 * avgRetentionAtPeriod returns the average retention across all cohorts at a given period.
 */
export function avgRetentionAtPeriod(cohorts: CohortRow[], period: number): number {
  const valid = cohorts.filter((c) => c.periods[period] !== undefined && c.periods[0] > 0);
  if (valid.length === 0) return 0;
  const sum = valid.reduce((acc, c) => acc + retentionRate(c, period), 0);
  return Math.round((sum / valid.length) * 10) / 10;
}

/**
 * cohortLabels returns formatted labels for cohort rows.
 */
export function cohortLabels(cohorts: CohortRow[]): string[] {
  return cohorts.map((c) => c.cohortDate);
}

/**
 * bestRetentionCohort returns the cohort with the highest period-1 retention.
 * Returns undefined for empty input.
 */
export function bestRetentionCohort(cohorts: CohortRow[]): CohortRow | undefined {
  if (cohorts.length === 0) return undefined;
  return cohorts.reduce((best, c) => (retentionRate(c, 1) > retentionRate(best, 1) ? c : best));
}
