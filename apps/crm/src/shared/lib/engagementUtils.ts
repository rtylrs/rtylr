export interface EngagementScore {
  recency: number;
  frequency: number;
  monetary: number;
  total: number;
  tier: 'champions' | 'loyal' | 'at-risk' | 'lost';
}

export function recencyScore(daysSinceVisit: number): number {
  if (daysSinceVisit <= 7) return 5;
  if (daysSinceVisit <= 30) return 4;
  if (daysSinceVisit <= 90) return 3;
  if (daysSinceVisit <= 180) return 2;
  return 1;
}

export function frequencyScore(visits: number): number {
  if (visits >= 20) return 5;
  if (visits >= 10) return 4;
  if (visits >= 5) return 3;
  if (visits >= 2) return 2;
  return 1;
}

export function monetaryScore(totalSpent: number): number {
  if (totalSpent >= 1000) return 5;
  if (totalSpent >= 500) return 4;
  if (totalSpent >= 200) return 3;
  if (totalSpent >= 50) return 2;
  return 1;
}

export function engagementScore(
  daysSinceVisit: number,
  visits: number,
  totalSpent: number
): EngagementScore {
  const recency = recencyScore(daysSinceVisit);
  const frequency = frequencyScore(visits);
  const monetary = monetaryScore(totalSpent);
  const total = recency + frequency + monetary;
  let tier: EngagementScore['tier'];
  if (total >= 13) tier = 'champions';
  else if (total >= 9) tier = 'loyal';
  else if (total >= 6) tier = 'at-risk';
  else tier = 'lost';
  return { recency, frequency, monetary, total, tier };
}

export const TIER_LABELS: Record<EngagementScore['tier'], string> = {
  champions: 'Champions',
  loyal: 'Loyal Customers',
  'at-risk': 'At Risk',
  lost: 'Lost',
};

export function tierLabel(tier: EngagementScore['tier']): string {
  return TIER_LABELS[tier];
}

export function isHighValue(score: EngagementScore): boolean {
  return score.tier === 'champions' || score.tier === 'loyal';
}

export function needsWinback(score: EngagementScore): boolean {
  return score.tier === 'at-risk' || score.tier === 'lost';
}
