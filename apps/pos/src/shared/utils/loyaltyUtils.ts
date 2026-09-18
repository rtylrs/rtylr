// Pure helpers for loyalty program calculations.
// These are extracted from loyaltyWorkflow for reuse across components.

/** Default exchange rate: 1 point per dollar spent. */
export const DEFAULT_POINTS_PER_DOLLAR = 1;

/** Default redemption rate: each point is worth $0.01. */
export const DEFAULT_POINT_VALUE = 0.01;

/** Minimum order total required to earn loyalty points. */
export const MIN_EARNABLE_TOTAL = 0.01;

/** Calculate points earned for a given order total. */
export function pointsEarned(
  totalAmount: number,
  pointsPerDollar = DEFAULT_POINTS_PER_DOLLAR,
): number {
  if (!Number.isFinite(totalAmount) || totalAmount < MIN_EARNABLE_TOTAL) return 0;
  if (!Number.isFinite(pointsPerDollar) || pointsPerDollar <= 0) return 0;
  return Math.floor(totalAmount * pointsPerDollar);
}

/** Calculate the discount amount from redeeming a number of points. */
export function pointsToDiscount(
  points: number,
  pointValue = DEFAULT_POINT_VALUE,
): number {
  if (!Number.isFinite(points) || points <= 0) return 0;
  if (!Number.isFinite(pointValue) || pointValue <= 0) return 0;
  return Math.round(points * pointValue * 100) / 100;
}

/** Calculate points required to get a specific discount amount. */
export function discountToPoints(
  discountAmount: number,
  pointValue = DEFAULT_POINT_VALUE,
): number {
  if (!Number.isFinite(discountAmount) || discountAmount <= 0) return 0;
  if (!Number.isFinite(pointValue) || pointValue <= 0) return 0;
  return Math.ceil(discountAmount / pointValue);
}

/** Cap redeemable points so the discount does not exceed the order total. */
export function cappedRedeemablePoints(
  availablePoints: number,
  orderTotal: number,
  pointValue = DEFAULT_POINT_VALUE,
): number {
  if (!Number.isFinite(availablePoints) || availablePoints <= 0) return 0;
  if (!Number.isFinite(orderTotal) || orderTotal <= 0) return 0;
  const maxByTotal = Math.floor(orderTotal / pointValue);
  return Math.min(availablePoints, maxByTotal);
}

/** Format a loyalty point balance for display. */
export function formatPoints(points: number): string {
  if (!Number.isFinite(points) || points < 0) return '0 pts';
  return `${Math.floor(points)} pts`;
}
