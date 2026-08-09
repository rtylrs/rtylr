// Operations analytics helpers for the insights operations feature.

import type {
  ShiftMetricsResponse,
  UtilizationResponse,
  TimeOffResponse,
} from '../hooks/queries/types';

/**
 * noShowRate returns the percentage of scheduled shifts that were no-shows.
 * Returns 0 when scheduled is 0.
 */
export function noShowRate(metrics: ShiftMetricsResponse): number {
  if (metrics.scheduled === 0) return 0;
  return (metrics.noShows / metrics.scheduled) * 100;
}

/**
 * completionRate returns the percentage of scheduled shifts completed.
 */
export function completionRate(metrics: ShiftMetricsResponse): number {
  return metrics.completionRatePct;
}

/**
 * missingShifts returns the number of scheduled shifts not yet completed.
 */
export function missingShifts(metrics: ShiftMetricsResponse): number {
  return metrics.scheduled - metrics.completed - metrics.noShows;
}

/**
 * unusedHours returns the number of available hours not utilized.
 */
export function unusedHours(utilization: UtilizationResponse): number {
  return (
    utilization.scheduledHours - utilization.availableHours * (utilization.utilizationPct / 100)
  );
}

/**
 * pendingTimeOffRatio returns pending requests as a fraction of total.
 * Returns 0 when there are no requests.
 */
export function pendingTimeOffRatio(timeOff: TimeOffResponse): number {
  const total = timeOff.pending + timeOff.approved;
  if (total === 0) return 0;
  return timeOff.pending / total;
}

/**
 * timeOffApprovalRate returns the percentage of time off requests approved.
 */
export function timeOffApprovalRate(timeOff: TimeOffResponse): number {
  const total = timeOff.pending + timeOff.approved;
  if (total === 0) return 0;
  return (timeOff.approved / total) * 100;
}
