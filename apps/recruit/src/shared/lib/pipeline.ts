// Recruitment pipeline helpers for the admin dashboard.
// The pipeline tracks submissions through statuses and produces
// funnel metrics (applications -> reviews -> shortlists -> hires).

import type { Submission } from './types';

export type PipelineStage = {
  status: string;
  label: string;
  count: number;
};

const PIPELINE_ORDER = ['pending', 'reviewing', 'shortlisted', 'hired'] as const;
const PIPELINE_LABELS: Record<string, string> = {
  pending: 'Applied',
  reviewing: 'In Review',
  shortlisted: 'Shortlisted',
  hired: 'Hired',
};

/**
 * buildPipeline returns an ordered array of pipeline stages with their
 * submission counts. Rejected submissions are excluded from the funnel
 * because they represent drop-offs, not progression.
 */
export function buildPipeline(submissions: Submission[]): PipelineStage[] {
  const counts = new Map<string, number>();
  for (const s of submissions) {
    if (s.status !== 'rejected') {
      counts.set(s.status, (counts.get(s.status) ?? 0) + 1);
    }
  }
  return PIPELINE_ORDER.map((status) => ({
    status,
    label: PIPELINE_LABELS[status],
    count: counts.get(status) ?? 0,
  }));
}

/**
 * pipelineConversionRates returns the percentage of candidates that
 * progressed from each stage to the next. Each element represents the
 * rate from stage[i] to stage[i+1].
 */
export function pipelineConversionRates(stages: PipelineStage[]): number[] {
  const rates: number[] = [];
  for (let i = 0; i < stages.length - 1; i++) {
    const from = stages[i].count;
    const to = stages[i + 1].count;
    rates.push(from === 0 ? 0 : (to / from) * 100);
  }
  return rates;
}

/**
 * overallConversionRate returns the percentage of applications that
 * resulted in a hire.
 */
export function overallConversionRate(submissions: Submission[]): number {
  if (submissions.length === 0) return 0;
  const hired = submissions.filter((s) => s.status === 'hired').length;
  return (hired / submissions.length) * 100;
}

/**
 * avgTimeToStatus returns the average number of days from submission
 * creation to reaching the given status. Submissions that have not
 * reached the target status are excluded.
 */
export function avgTimeToStatus(submissions: Submission[], targetStatus: string): number {
  const matching = submissions.filter((s) => s.status === targetStatus);
  if (matching.length === 0) return 0;
  const totalDays = matching.reduce((sum, s) => {
    const created = new Date(s.createdAt).getTime();
    const updated = new Date(s.updatedAt).getTime();
    return sum + (updated - created) / (1000 * 60 * 60 * 24);
  }, 0);
  return totalDays / matching.length;
}
