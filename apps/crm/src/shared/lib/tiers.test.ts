import { describe, it, expect } from 'vitest';
import {
  getCustomerTier,
  getTier,
  nextTier,
  pointsToNextTier,
  tierColor,
  tierThresholds,
  thresholdsFromConfig,
} from './tiers';
import type { Tier } from './tiers';

// ── getTier ───────────────────────────────────────────────────────────────────

describe('getTier', () => {
  it('returns Bronze for 0 points', () => {
    expect(getTier(0)).toBe('Bronze');
  });

  it('returns Bronze for points below the Silver threshold', () => {
    expect(getTier(1)).toBe('Bronze');
    expect(getTier(499)).toBe('Bronze');
  });

  it('returns Silver at exactly 500 points', () => {
    expect(getTier(500)).toBe('Silver');
  });

  it('returns Silver for points between Silver and Gold thresholds', () => {
    expect(getTier(501)).toBe('Silver');
    expect(getTier(999)).toBe('Silver');
  });

  it('returns Gold at exactly 1000 points', () => {
    expect(getTier(1000)).toBe('Gold');
  });

  it('returns Gold for points between Gold and Platinum thresholds', () => {
    expect(getTier(1001)).toBe('Gold');
    expect(getTier(2499)).toBe('Gold');
  });

  it('returns Platinum at exactly 2500 points', () => {
    expect(getTier(2500)).toBe('Platinum');
  });

  it('returns Platinum for points above 2500', () => {
    expect(getTier(2501)).toBe('Platinum');
    expect(getTier(9999)).toBe('Platinum');
  });
});

describe('persisted/configured tiers', () => {
  it('prefers a persisted backend tier over stale local thresholds', () => {
    expect(getCustomerTier(600, 'gold')).toBe('Gold');
  });

  it('uses configured thresholds for customers without a persisted tier', () => {
    const thresholds = thresholdsFromConfig({ silverMin: 100, goldMin: 300, platinumMin: 700 });
    expect(getTier(350, thresholds)).toBe('Gold');
    expect(pointsToNextTier(350, thresholds)).toBe(350);
  });
});

// ── nextTier ──────────────────────────────────────────────────────────────────

describe('nextTier', () => {
  it('Bronze advances to Silver', () => {
    expect(nextTier('Bronze')).toBe('Silver');
  });

  it('Silver advances to Gold', () => {
    expect(nextTier('Silver')).toBe('Gold');
  });

  it('Gold advances to Platinum', () => {
    expect(nextTier('Gold')).toBe('Platinum');
  });

  it('Platinum has no next tier (returns null)', () => {
    expect(nextTier('Platinum')).toBeNull();
  });
});

// ── pointsToNextTier ──────────────────────────────────────────────────────────

describe('pointsToNextTier', () => {
  it('returns the gap to Silver from 0 Bronze points', () => {
    expect(pointsToNextTier(0)).toBe(500);
  });

  it('returns the correct remaining gap mid-Bronze', () => {
    expect(pointsToNextTier(200)).toBe(300);
  });

  it('returns 0 when already exactly at the Silver threshold', () => {
    // At 500 points the customer IS Silver, so next is Gold (1000 pts).
    expect(pointsToNextTier(500)).toBe(500);
  });

  it('returns the gap to Gold from a Silver member', () => {
    expect(pointsToNextTier(750)).toBe(250);
  });

  it('returns the gap to Platinum from a Gold member', () => {
    expect(pointsToNextTier(1000)).toBe(1500);
    expect(pointsToNextTier(1500)).toBe(1000);
  });

  it('returns null for Platinum members (already max tier)', () => {
    expect(pointsToNextTier(2500)).toBeNull();
    expect(pointsToNextTier(9999)).toBeNull();
  });
});

// ── tierColor & tierThresholds dictionaries ───────────────────────────────────

describe('tierColor', () => {
  const allTiers: Tier[] = ['Bronze', 'Silver', 'Gold', 'Platinum'];

  it('has a color entry for every tier', () => {
    for (const tier of allTiers) {
      expect(tierColor[tier]).toBeDefined();
      expect(typeof tierColor[tier]).toBe('string');
    }
  });

  it('color values are non-empty strings', () => {
    for (const tier of allTiers) {
      expect(tierColor[tier].length).toBeGreaterThan(0);
    }
  });
});

describe('tierThresholds', () => {
  it('Bronze threshold is 0 (starting tier)', () => {
    expect(tierThresholds.Bronze).toBe(0);
  });

  it('thresholds are strictly ascending', () => {
    expect(tierThresholds.Bronze).toBeLessThan(tierThresholds.Silver);
    expect(tierThresholds.Silver).toBeLessThan(tierThresholds.Gold);
    expect(tierThresholds.Gold).toBeLessThan(tierThresholds.Platinum);
  });

  it('getTier is consistent with tierThresholds boundaries', () => {
    // At each threshold boundary the correct tier should be returned.
    const tiers: Tier[] = ['Bronze', 'Silver', 'Gold', 'Platinum'];
    for (const tier of tiers) {
      expect(getTier(tierThresholds[tier])).toBe(tier);
    }
  });
});
