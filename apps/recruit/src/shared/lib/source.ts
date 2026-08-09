// Candidate source tracking helpers for the recruit admin panel.

export type CandidateSource =
  | 'job_board'
  | 'referral'
  | 'linkedin'
  | 'agency'
  | 'direct'
  | 'campus'
  | 'other';

export interface SourceMetric {
  source: CandidateSource;
  applications: number;
  hired: number;
  avgDaysToHire: number;
}

/**
 * sourceLabel returns a display-friendly label for a candidate source.
 */
export function sourceLabel(source: CandidateSource | string): string {
  const labels: Record<string, string> = {
    job_board: 'Job Board',
    referral: 'Referral',
    linkedin: 'LinkedIn',
    agency: 'Agency',
    direct: 'Direct Application',
    campus: 'Campus Recruiting',
    other: 'Other',
  };
  return labels[source] ?? source;
}

/**
 * hireRate returns the percentage of applications that resulted in a hire.
 */
export function hireRate(metric: SourceMetric): number {
  if (metric.applications === 0) return 0;
  return Math.round((metric.hired / metric.applications) * 1000) / 10;
}

/**
 * topSourceByHires returns the source with the most hires.
 */
export function topSourceByHires(metrics: SourceMetric[]): SourceMetric | undefined {
  if (metrics.length === 0) return undefined;
  return metrics.reduce((best, m) => (m.hired > best.hired ? m : best));
}

/**
 * fastestHiringSource returns the source with the lowest average days to hire.
 */
export function fastestHiringSource(metrics: SourceMetric[]): SourceMetric | undefined {
  const valid = metrics.filter((m) => m.hired > 0);
  if (valid.length === 0) return undefined;
  return valid.reduce((best, m) => (m.avgDaysToHire < best.avgDaysToHire ? m : best));
}

/**
 * totalApplications returns the sum of applications across all sources.
 */
export function totalApplications(metrics: SourceMetric[]): number {
  return metrics.reduce((sum, m) => sum + m.applications, 0);
}

/**
 * allCandidateSources returns all defined candidate source values.
 */
export function allCandidateSources(): CandidateSource[] {
  return ['job_board', 'referral', 'linkedin', 'agency', 'direct', 'campus', 'other'];
}
