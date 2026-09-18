// Net Promoter Score helpers for the insights analytics feature.

export type NpsCategory = 'promoter' | 'passive' | 'detractor';

export interface NpsResponse {
  respondentId: string;
  score: number;
  comment?: string;
  submittedAt: string;
}

/**
 * npsCategory classifies a score into promoter (9-10), passive (7-8), or detractor (0-6).
 */
export function npsCategory(score: number): NpsCategory {
  if (score >= 9) return 'promoter';
  if (score >= 7) return 'passive';
  return 'detractor';
}

/**
 * npsScore calculates the Net Promoter Score from a set of responses.
 * NPS = % Promoters - % Detractors
 */
export function npsScore(responses: NpsResponse[]): number {
  if (responses.length === 0) return 0;
  const promoters = responses.filter((r) => npsCategory(r.score) === 'promoter').length;
  const detractors = responses.filter((r) => npsCategory(r.score) === 'detractor').length;
  const pct = (v: number) => (v / responses.length) * 100;
  return Math.round(pct(promoters) - pct(detractors));
}

/**
 * npsCategoryCount returns the count for each NPS category.
 */
export function npsCategoryCount(responses: NpsResponse[]): Record<NpsCategory, number> {
  const counts: Record<NpsCategory, number> = { promoter: 0, passive: 0, detractor: 0 };
  for (const r of responses) {
    counts[npsCategory(r.score)]++;
  }
  return counts;
}

/**
 * npsLabel returns a qualitative label based on NPS score.
 */
export function npsLabel(score: number): string {
  if (score >= 50) return 'Excellent';
  if (score >= 30) return 'Good';
  if (score >= 0) return 'Neutral';
  return 'Poor';
}

/**
 * avgNpsScore returns the average raw score across all responses.
 */
export function avgNpsScore(responses: NpsResponse[]): number {
  if (responses.length === 0) return 0;
  const sum = responses.reduce((acc, r) => acc + r.score, 0);
  return Math.round((sum / responses.length) * 10) / 10;
}
