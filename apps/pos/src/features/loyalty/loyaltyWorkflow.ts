import { createOneKitClient } from '@voxire/api';
import { LoyaltyServiceClient } from '@voxire/voxie/pos/loyalty/v1/client';

const loyaltyClient = createOneKitClient(LoyaltyServiceClient, 'pos');

// Points-per-dollar earned (configurable via /pos/loyalty/config, default 1 point per $1)
export const DEFAULT_POINTS_PER_DOLLAR = 1;
// Dollar value per point when redeeming (default 1 point = $0.01)
export const DEFAULT_POINT_VALUE = 0.01;

export interface EarnResult {
  earned: number;
  newTotal: number | null;
}

export interface RedeemResult {
  pointsUsed: number;
  discountAmount: number;
}

/** How many points a purchase earns given a total amount. */
export function calculateEarnedPoints(
  totalAmount: number,
  pointsPerDollar = DEFAULT_POINTS_PER_DOLLAR,
): number {
  if (!Number.isFinite(totalAmount) || totalAmount <= 0) return 0;
  if (!Number.isFinite(pointsPerDollar) || pointsPerDollar <= 0) return 0;
  return Math.floor(totalAmount * pointsPerDollar);
}

/** Dollar discount value of redeeming a given number of points. */
export function calculateRedemptionDiscount(
  pointsToRedeem: number,
  pointValue = DEFAULT_POINT_VALUE,
): number {
  if (!Number.isFinite(pointsToRedeem) || pointsToRedeem <= 0) return 0;
  if (!Number.isFinite(pointValue) || pointValue <= 0) return 0;
  return Math.round(pointsToRedeem * pointValue * 100) / 100;
}

/** Max redeemable points capped so discount doesn't exceed the order total. */
export function maxRedeemablePoints(
  availablePoints: number,
  orderTotal: number,
  pointValue = DEFAULT_POINT_VALUE,
): number {
  const maxByBalance = availablePoints;
  const maxByTotal = Math.floor(orderTotal / pointValue);
  return Math.min(maxByBalance, maxByTotal);
}

/** Call backend to earn points after a successful payment. No-op if no customer. */
export async function earnPointsAfterPayment(params: {
  customerId: number | null;
  orderId: number;
  totalAmount: number;
  pointsPerDollar?: number;
}): Promise<EarnResult | null> {
  const { customerId, orderId, totalAmount, pointsPerDollar } = params;
  if (!customerId || customerId <= 0) return null;

  const earned = calculateEarnedPoints(totalAmount, pointsPerDollar);
  if (earned === 0) return { earned: 0, newTotal: null };

  await loyaltyClient.earnPoints({
    customerId: String(customerId),
    points: earned,
    orderId: String(orderId),
  });
  return { earned, newTotal: null };
}

/** Call backend to redeem points as a discount before checkout. No-op if no customer. */
export async function redeemPointsForDiscount(params: {
  customerId: number | null;
  orderId: number;
  pointsToRedeem: number;
  pointValue?: number;
}): Promise<RedeemResult | null> {
  const { customerId, orderId, pointsToRedeem, pointValue } = params;
  if (!customerId || customerId <= 0) return null;
  if (pointsToRedeem <= 0) return null;

  const discountAmount = calculateRedemptionDiscount(pointsToRedeem, pointValue);
  await loyaltyClient.redeemPoints({
    customerId: String(customerId),
    points: pointsToRedeem,
  });
  return { pointsUsed: pointsToRedeem, discountAmount };
}
