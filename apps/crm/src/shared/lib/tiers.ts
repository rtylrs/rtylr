export type Tier = 'Bronze' | 'Silver' | 'Gold' | 'Platinum';

export const tierColor: Record<Tier, string> = {
  Platinum: '#e879f9',
  Gold: '#f59e0b',
  Silver: '#94a3b8',
  Bronze: '#b45309',
};

export const tierThresholds: Record<Tier, number> = {
  Bronze: 0,
  Silver: 500,
  Gold: 1000,
  Platinum: 2500,
};

export function getTier(points: number): Tier {
  if (points >= 2500) return 'Platinum';
  if (points >= 1000) return 'Gold';
  if (points >= 500) return 'Silver';
  return 'Bronze';
}

export function nextTier(tier: Tier): Tier | null {
  if (tier === 'Bronze') return 'Silver';
  if (tier === 'Silver') return 'Gold';
  if (tier === 'Gold') return 'Platinum';
  return null;
}

export function pointsToNextTier(points: number): number | null {
  const cur = getTier(points);
  const next = nextTier(cur);
  if (!next) return null;
  return tierThresholds[next] - points;
}
