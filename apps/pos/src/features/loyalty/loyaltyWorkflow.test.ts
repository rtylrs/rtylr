import { describe, expect, it, vi, beforeEach } from 'vitest';
import {
  calculateEarnedPoints,
  calculateRedemptionDiscount,
  maxRedeemablePoints,
  earnPointsAfterPayment,
  redeemPointsForDiscount,
} from './loyaltyWorkflow';

vi.mock('../../shared/services/api/posService', () => ({
  posService: {
    earnLoyaltyPoints: vi.fn(),
    redeemLoyaltyPoints: vi.fn(),
  },
}));

vi.mock('@voxire/api', () => ({
  unwrap: vi.fn((p) => p),
}));

import { posService } from '../../shared/services/api/posService';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('calculateEarnedPoints', () => {
  it('earns 1 point per dollar by default', () => {
    expect(calculateEarnedPoints(25)).toBe(25);
  });

  it('floors fractional points', () => {
    expect(calculateEarnedPoints(9.99)).toBe(9);
  });

  it('returns 0 for non-positive amounts', () => {
    expect(calculateEarnedPoints(0)).toBe(0);
    expect(calculateEarnedPoints(-5)).toBe(0);
  });

  it('scales with custom points-per-dollar rate', () => {
    expect(calculateEarnedPoints(10, 2)).toBe(20);
    expect(calculateEarnedPoints(10, 0.5)).toBe(5);
  });
});

describe('calculateRedemptionDiscount', () => {
  it('converts 100 points to $1.00 at default rate', () => {
    expect(calculateRedemptionDiscount(100)).toBe(1.0);
  });

  it('rounds to 2 decimal places', () => {
    expect(calculateRedemptionDiscount(333)).toBeCloseTo(3.33, 2);
  });

  it('returns 0 for zero or negative points', () => {
    expect(calculateRedemptionDiscount(0)).toBe(0);
    expect(calculateRedemptionDiscount(-10)).toBe(0);
  });
});

describe('maxRedeemablePoints', () => {
  it('caps at available balance when balance is limiting factor', () => {
    // 50 points = $0.50, order is $10 - balance is the limit
    expect(maxRedeemablePoints(50, 10)).toBe(50);
  });

  it('caps at order total when total is the limiting factor', () => {
    // 1000 points = $10.00, order is $5 - 500 points max
    expect(maxRedeemablePoints(1000, 5)).toBe(500);
  });

  it('handles zero order total', () => {
    expect(maxRedeemablePoints(100, 0)).toBe(0);
  });
});

describe('earnPointsAfterPayment', () => {
  it('calls earnLoyaltyPoints with correct points count', async () => {
    vi.mocked(posService.earnLoyaltyPoints).mockResolvedValue(undefined as never);

    const result = await earnPointsAfterPayment({
      customerId: 42,
      orderId: 500,
      totalAmount: 25.0,
    });

    expect(result).toEqual({ earned: 25, newTotal: null });
    expect(posService.earnLoyaltyPoints).toHaveBeenCalledWith({
      customer_id: 42,
      points: 25,
      order_id: 500,
    });
  });

  it('returns null and skips API call when no customer attached', async () => {
    const result = await earnPointsAfterPayment({
      customerId: null,
      orderId: 500,
      totalAmount: 25.0,
    });

    expect(result).toBeNull();
    expect(posService.earnLoyaltyPoints).not.toHaveBeenCalled();
  });

  it('skips API call when total earns 0 points', async () => {
    const result = await earnPointsAfterPayment({
      customerId: 1,
      orderId: 1,
      totalAmount: 0,
    });

    expect(result).toEqual({ earned: 0, newTotal: null });
    expect(posService.earnLoyaltyPoints).not.toHaveBeenCalled();
  });
});

describe('redeemPointsForDiscount', () => {
  it('calls redeemLoyaltyPoints and returns discount amount', async () => {
    vi.mocked(posService.redeemLoyaltyPoints).mockResolvedValue(undefined as never);

    const result = await redeemPointsForDiscount({
      customerId: 42,
      orderId: 500,
      pointsToRedeem: 200,
    });

    expect(result).toEqual({ pointsUsed: 200, discountAmount: 2.0 });
    expect(posService.redeemLoyaltyPoints).toHaveBeenCalledWith({
      customer_id: 42,
      points: 200,
      order_id: 500,
    });
  });

  it('returns null and skips API when no customer attached', async () => {
    const result = await redeemPointsForDiscount({
      customerId: null,
      orderId: 500,
      pointsToRedeem: 100,
    });

    expect(result).toBeNull();
    expect(posService.redeemLoyaltyPoints).not.toHaveBeenCalled();
  });

  it('returns null and skips API when points to redeem is 0', async () => {
    const result = await redeemPointsForDiscount({
      customerId: 42,
      orderId: 500,
      pointsToRedeem: 0,
    });

    expect(result).toBeNull();
    expect(posService.redeemLoyaltyPoints).not.toHaveBeenCalled();
  });
});
