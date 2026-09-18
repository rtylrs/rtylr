// Business scorecard helpers for the insights analytics feature.

export type ScorecardRating = 'excellent' | 'good' | 'average' | 'poor';

export interface ScorecardMetric {
  name: string;
  value: number;
  target: number;
  weight: number;
}

/**
 * metricScore returns the percentage score for a single metric.
 */
export function metricScore(metric: ScorecardMetric): number {
  if (metric.target === 0) return 0;
  return Math.min(100, Math.round((metric.value / metric.target) * 1000) / 10);
}

/**
 * weightedScore calculates the weighted aggregate score across all metrics.
 */
export function weightedScore(metrics: ScorecardMetric[]): number {
  const totalWeight = metrics.reduce((sum, m) => sum + m.weight, 0);
  if (totalWeight === 0) return 0;
  const sum = metrics.reduce((acc, m) => acc + metricScore(m) * m.weight, 0);
  return Math.round((sum / totalWeight) * 10) / 10;
}

/**
 * scorecardRating returns a qualitative rating based on the weighted score.
 */
export function scorecardRating(score: number): ScorecardRating {
  if (score >= 90) return 'excellent';
  if (score >= 70) return 'good';
  if (score >= 50) return 'average';
  return 'poor';
}

/**
 * ratingLabel returns a display-friendly label for a scorecard rating.
 */
export function ratingLabel(rating: ScorecardRating): string {
  const labels: Record<ScorecardRating, string> = {
    excellent: 'Excellent',
    good: 'Good',
    average: 'Average',
    poor: 'Poor',
  };
  return labels[rating];
}

/**
 * underPerformingMetrics returns metrics scoring below the given threshold.
 */
export function underPerformingMetrics(
  metrics: ScorecardMetric[],
  threshold = 70
): ScorecardMetric[] {
  return metrics.filter((m) => metricScore(m) < threshold);
}
