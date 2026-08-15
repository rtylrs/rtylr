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

export interface TierThresholdConfig {
  bronzeMin?: number;
  silverMin?: number;
  goldMin?: number;
  platinumMin?: number;
}

export function thresholdsFromConfig(config?: TierThresholdConfig): Record<Tier, number> {
  const value = (candidate: number | undefined, fallback: number): number =>
    typeof candidate === 'number' && Number.isFinite(candidate) ? Math.max(0, candidate) : fallback;
  return {
    Bronze: value(config?.bronzeMin, tierThresholds.Bronze),
    Silver: value(config?.silverMin, tierThresholds.Silver),
    Gold: value(config?.goldMin, tierThresholds.Gold),
    Platinum: value(config?.platinumMin, tierThresholds.Platinum),
  };
}

export function normalizeTier(value?: string | null): Tier | undefined {
  const normalized = value?.trim().toLowerCase();
  if (normalized === 'bronze') return 'Bronze';
  if (normalized === 'silver') return 'Silver';
  if (normalized === 'gold') return 'Gold';
  if (normalized === 'platinum') return 'Platinum';
  return undefined;
}

export function getTier(points: number, thresholds: Record<Tier, number> = tierThresholds): Tier {
  if (points >= thresholds.Platinum) return 'Platinum';
  if (points >= thresholds.Gold) return 'Gold';
  if (points >= thresholds.Silver) return 'Silver';
  return 'Bronze';
}

/** Prefer the backend's persisted tier; fall back to thresholds for legacy payloads. */
export function getCustomerTier(
  points: number,
  storedTier?: string | null,
  thresholds: Record<Tier, number> = tierThresholds
): Tier {
  return normalizeTier(storedTier) ?? getTier(points, thresholds);
}

export function nextTier(tier: Tier): Tier | null {
  if (tier === 'Bronze') return 'Silver';
  if (tier === 'Silver') return 'Gold';
  if (tier === 'Gold') return 'Platinum';
  return null;
}

export function pointsToNextTier(
  points: number,
  thresholds: Record<Tier, number> = tierThresholds,
  storedTier?: string | null
): number | null {
  const cur = getCustomerTier(points, storedTier, thresholds);
  const next = nextTier(cur);
  if (!next) return null;
  return Math.max(0, thresholds[next] - points);
}
