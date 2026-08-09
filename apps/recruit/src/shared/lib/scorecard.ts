// Candidate scorecard helpers for the recruit admin panel.

export interface ScorecardCriteria {
  id: string;
  name: string;
  weight: number;
  score: number;
  maxScore: number;
}

/**
 * criteriaScore returns the percentage score for a single criterion.
 */
export function criteriaScore(criteria: ScorecardCriteria): number {
  if (criteria.maxScore === 0) return 0;
  return Math.round((criteria.score / criteria.maxScore) * 1000) / 10;
}

/**
 * weightedCandidateScore returns the weighted total score across all criteria.
 */
export function weightedCandidateScore(criteria: ScorecardCriteria[]): number {
  const totalWeight = criteria.reduce((sum, c) => sum + c.weight, 0);
  if (totalWeight === 0) return 0;
  const sum = criteria.reduce((acc, c) => acc + criteriaScore(c) * c.weight, 0);
  return Math.round((sum / totalWeight) * 10) / 10;
}

/**
 * passesThreshold returns true when the weighted score meets the minimum threshold.
 */
export function passesThreshold(criteria: ScorecardCriteria[], threshold: number): boolean {
  return weightedCandidateScore(criteria) >= threshold;
}

/**
 * weakCriteria returns criteria scoring below a given percentage threshold.
 */
export function weakCriteria(criteria: ScorecardCriteria[], threshold = 60): ScorecardCriteria[] {
  return criteria.filter((c) => criteriaScore(c) < threshold);
}

/**
 * topCriteria returns criteria scoring at or above a given percentage threshold.
 */
export function topCriteria(criteria: ScorecardCriteria[], threshold = 80): ScorecardCriteria[] {
  return criteria.filter((c) => criteriaScore(c) >= threshold);
}

/**
 * scorecardSummary returns a brief summary string for the scorecard result.
 */
export function scorecardSummary(criteria: ScorecardCriteria[], threshold = 70): string {
  const score = weightedCandidateScore(criteria);
  const passes = score >= threshold;
  return passes ? `Recommended (${score}%)` : `Not Recommended (${score}%)`;
}
