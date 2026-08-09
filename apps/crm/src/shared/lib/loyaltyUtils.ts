import type { LoyaltyMember, LoyaltyTier } from '../schemas/loyalty';

/**
 * Returns the highest tier whose threshold the member meets.
 * Tiers should be sorted by threshold ascending.
 */
export function getMemberTier(member: LoyaltyMember, tiers: LoyaltyTier[]): LoyaltyTier | null {
  const sorted = [...tiers].sort((a, b) => a.threshold - b.threshold);
  let current: LoyaltyTier | null = null;
  for (const tier of sorted) {
    if (member.points >= tier.threshold) {
      current = tier;
    }
  }
  return current;
}

/**
 * Returns the next tier above the member's current tier, or null if at max.
 */
export function getNextTier(member: LoyaltyMember, tiers: LoyaltyTier[]): LoyaltyTier | null {
  const sorted = [...tiers].sort((a, b) => a.threshold - b.threshold);
  for (const tier of sorted) {
    if (member.points < tier.threshold) {
      return tier;
    }
  }
  return null;
}

/**
 * Returns the number of points needed to reach the next tier,
 * or null if the member is already at the highest tier.
 */
export function pointsToNextTier(member: LoyaltyMember, tiers: LoyaltyTier[]): number | null {
  const next = getNextTier(member, tiers);
  if (!next) return null;
  return next.threshold - member.points;
}

/**
 * Returns the earn multiplier for the member's current tier.
 * Falls back to 1.0 if no tier matches.
 */
export function getEarnMultiplier(member: LoyaltyMember, tiers: LoyaltyTier[]): number {
  const tier = getMemberTier(member, tiers);
  return tier?.multiplier ?? 1.0;
}

/**
 * Returns true when the member has enough points to redeem the given amount.
 */
export function canRedeem(member: LoyaltyMember, requiredPoints: number): boolean {
  return member.points >= requiredPoints;
}

/**
 * Formats a points number with comma separators and a "pts" suffix.
 * e.g. 1200 -> "1,200 pts"
 */
export function formatPoints(points: number): string {
  return `${points.toLocaleString('en-US')} pts`;
}
