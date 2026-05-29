// Diversity and inclusion metrics helpers for the recruit admin panel.

export interface DiversityMetric {
  category: string;
  group: string;
  count: number;
}

/**
 * totalCount returns the total count across all diversity metrics.
 */
export function totalCount(metrics: DiversityMetric[]): number {
  return metrics.reduce((sum, m) => sum + m.count, 0);
}

/**
 * groupShare returns the percentage share for a specific group.
 */
export function groupShare(metric: DiversityMetric, metrics: DiversityMetric[]): number {
  const total = totalCount(metrics);
  if (total === 0) return 0;
  return Math.round((metric.count / total) * 1000) / 10;
}

/**
 * diversityIndex returns a simple diversity score (1 minus sum of squared proportions).
 * Returns 0 for empty or single-group data.
 */
export function diversityIndex(metrics: DiversityMetric[]): number {
  const total = totalCount(metrics);
  if (total === 0 || metrics.length <= 1) return 0;
  const sumSq = metrics.reduce((sum, m) => {
    const p = m.count / total;
    return sum + p * p;
  }, 0);
  return Math.round((1 - sumSq) * 1000) / 1000;
}

/**
 * topGroup returns the group with the highest count in a category.
 */
export function topGroup(metrics: DiversityMetric[]): DiversityMetric | undefined {
  if (metrics.length === 0) return undefined;
  return metrics.reduce((best, m) => (m.count > best.count ? m : best));
}

/**
 * filterByCategory returns metrics filtered by category name.
 */
export function filterByCategory(metrics: DiversityMetric[], category: string): DiversityMetric[] {
  return metrics.filter((m) => m.category === category);
}

/**
 * categoryNames returns unique category names from a metrics array.
 */
export function categoryNames(metrics: DiversityMetric[]): string[] {
  return [...new Set(metrics.map((m) => m.category))];
}
