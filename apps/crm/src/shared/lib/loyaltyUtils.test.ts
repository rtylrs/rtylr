import { describe, it, expect } from 'vitest';
import {
  getMemberTier,
  getNextTier,
  pointsToNextTier,
  getEarnMultiplier,
  canRedeem,
  formatPoints,
} from './loyaltyUtils';
import type { LoyaltyMember, LoyaltyTier } from '../schemas/loyalty';

const tiers: LoyaltyTier[] = [
  { id: 1, name: 'Bronze', threshold: 0, multiplier: 1.0, perks: [], color: '#b45309' },
  { id: 2, name: 'Silver', threshold: 500, multiplier: 1.5, perks: [], color: '#94a3b8' },
  { id: 3, name: 'Gold', threshold: 1000, multiplier: 2.0, perks: [], color: '#f59e0b' },
  { id: 4, name: 'Platinum', threshold: 2500, multiplier: 3.0, perks: [], color: '#e879f9' },
];

const makeMember = (points: number): LoyaltyMember => ({
  id: 1,
  name: 'Jane Doe',
  initials: 'JD',
  points,
  tier: 'Bronze',
  visits: 10,
  joinedAt: '2023-01-01',
  lastActive: '2024-05-01',
});

describe('getMemberTier', () => {
  it('returns Bronze for 0 points', () => {
    expect(getMemberTier(makeMember(0), tiers)?.name).toBe('Bronze');
  });

  it('returns Silver for 500 points', () => {
    expect(getMemberTier(makeMember(500), tiers)?.name).toBe('Silver');
  });

  it('returns Gold for 1200 points', () => {
    expect(getMemberTier(makeMember(1200), tiers)?.name).toBe('Gold');
  });

  it('returns Platinum for 3000 points', () => {
    expect(getMemberTier(makeMember(3000), tiers)?.name).toBe('Platinum');
  });
});

describe('getNextTier', () => {
  it('returns Silver when member has 0 points', () => {
    expect(getNextTier(makeMember(0), tiers)?.name).toBe('Silver');
  });

  it('returns Gold when member is at Silver', () => {
    expect(getNextTier(makeMember(500), tiers)?.name).toBe('Gold');
  });

  it('returns null for Platinum members', () => {
    expect(getNextTier(makeMember(3000), tiers)).toBeNull();
  });
});

describe('pointsToNextTier', () => {
  it('returns 500 for a member with 0 points', () => {
    expect(pointsToNextTier(makeMember(0), tiers)).toBe(500);
  });

  it('returns 300 for a member with 700 points (Gold at 1000)', () => {
    expect(pointsToNextTier(makeMember(700), tiers)).toBe(300);
  });

  it('returns null for Platinum members', () => {
    expect(pointsToNextTier(makeMember(2500), tiers)).toBeNull();
  });
});

describe('getEarnMultiplier', () => {
  it('returns 1.0 for Bronze members', () => {
    expect(getEarnMultiplier(makeMember(0), tiers)).toBe(1.0);
  });

  it('returns 2.0 for Gold members', () => {
    expect(getEarnMultiplier(makeMember(1000), tiers)).toBe(2.0);
  });

  it('returns 3.0 for Platinum members', () => {
    expect(getEarnMultiplier(makeMember(2500), tiers)).toBe(3.0);
  });
});

describe('canRedeem', () => {
  it('returns true when member has enough points', () => {
    expect(canRedeem(makeMember(500), 200)).toBe(true);
  });

  it('returns false when member does not have enough points', () => {
    expect(canRedeem(makeMember(100), 200)).toBe(false);
  });

  it('returns true when member has exactly the required points', () => {
    expect(canRedeem(makeMember(200), 200)).toBe(true);
  });
});

describe('formatPoints', () => {
  it('formats points with comma separators and pts suffix', () => {
    expect(formatPoints(1200)).toBe('1,200 pts');
  });

  it('formats 0 correctly', () => {
    expect(formatPoints(0)).toBe('0 pts');
  });
});
