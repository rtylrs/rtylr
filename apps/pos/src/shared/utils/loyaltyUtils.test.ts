import { describe, it, expect } from 'vitest';
import {
  pointsEarned,
  pointsToDiscount,
  discountToPoints,
  cappedRedeemablePoints,
  formatPoints,
  DEFAULT_POINTS_PER_DOLLAR,
  DEFAULT_POINT_VALUE,
} from './loyaltyUtils';

describe('pointsEarned', () => {
  it('earns 1 point per dollar by default', () => {
    expect(pointsEarned(25)).toBe(25);
  });

  it('earns the correct number of points with a custom rate', () => {
    expect(pointsEarned(10, 2)).toBe(20);
  });

  it('floors fractional points', () => {
    expect(pointsEarned(10.5, DEFAULT_POINTS_PER_DOLLAR)).toBe(10);
  });

  it('returns 0 for zero total', () => {
    expect(pointsEarned(0)).toBe(0);
  });

  it('returns 0 for a negative total', () => {
    expect(pointsEarned(-5)).toBe(0);
  });

  it('returns 0 for a zero points-per-dollar rate', () => {
    expect(pointsEarned(50, 0)).toBe(0);
  });

  it('returns 0 for NaN inputs', () => {
    expect(pointsEarned(NaN)).toBe(0);
    expect(pointsEarned(50, NaN)).toBe(0);
  });
});

describe('pointsToDiscount', () => {
  it('converts 100 points to $1.00 at the default rate', () => {
    expect(pointsToDiscount(100)).toBe(1);
  });

  it('converts 50 points to $0.50', () => {
    expect(pointsToDiscount(50, DEFAULT_POINT_VALUE)).toBe(0.5);
  });

  it('returns 0 for zero points', () => {
    expect(pointsToDiscount(0)).toBe(0);
  });

  it('returns 0 for negative points', () => {
    expect(pointsToDiscount(-10)).toBe(0);
  });

  it('rounds to two decimal places', () => {
    expect(pointsToDiscount(1, 0.015)).toBeCloseTo(0.02, 2);
  });
});

describe('discountToPoints', () => {
  it('converts $1.00 to 100 points at the default rate', () => {
    expect(discountToPoints(1)).toBe(100);
  });

  it('ceils fractional point requirements', () => {
    expect(discountToPoints(0.5, 0.015)).toBe(34);
  });

  it('returns 0 for a zero discount', () => {
    expect(discountToPoints(0)).toBe(0);
  });

  it('returns 0 for a negative discount', () => {
    expect(discountToPoints(-1)).toBe(0);
  });
});

describe('cappedRedeemablePoints', () => {
  it('returns all available points when they are under the order cap', () => {
    expect(cappedRedeemablePoints(50, 10, 0.01)).toBe(50);
  });

  it('caps points to prevent the discount from exceeding the order total', () => {
    expect(cappedRedeemablePoints(2000, 10, 0.01)).toBe(1000);
  });

  it('returns 0 when no points are available', () => {
    expect(cappedRedeemablePoints(0, 10)).toBe(0);
  });

  it('returns 0 when order total is 0', () => {
    expect(cappedRedeemablePoints(100, 0)).toBe(0);
  });

  it('returns 0 for negative points', () => {
    expect(cappedRedeemablePoints(-10, 50)).toBe(0);
  });
});

describe('formatPoints', () => {
  it('formats a positive point balance', () => {
    expect(formatPoints(150)).toBe('150 pts');
  });

  it('formats zero points', () => {
    expect(formatPoints(0)).toBe('0 pts');
  });

  it('returns "0 pts" for negative points', () => {
    expect(formatPoints(-5)).toBe('0 pts');
  });

  it('returns "0 pts" for NaN', () => {
    expect(formatPoints(NaN)).toBe('0 pts');
  });

  it('floors fractional points', () => {
    expect(formatPoints(9.9)).toBe('9 pts');
  });
});
