import { describe, it, expect } from 'vitest';
import {
  referralsByStatus,
  totalBonusPaid,
  referralHireRate,
  topReferrers,
  activeReferrals,
} from './referral';
import type { Referral } from './referral';

function makeReferral(
  status: 'pending' | 'interviewing' | 'hired' | 'rejected',
  referrerId = 'r1',
  bonus?: number
): Referral {
  return {
    id: `ref-${Math.random()}`,
    referrerId,
    candidateName: 'John Doe',
    position: 'Engineer',
    status,
    submittedAt: '2026-01-01',
    bonusAmount: bonus,
  };
}

const referrals = [
  makeReferral('hired', 'emp1', 2000),
  makeReferral('hired', 'emp2', 1500),
  makeReferral('pending', 'emp1'),
  makeReferral('interviewing', 'emp3'),
  makeReferral('rejected', 'emp2'),
];

describe('referralsByStatus', () => {
  it('groups referrals by status', () => {
    const grouped = referralsByStatus(referrals);
    expect(grouped['hired']).toHaveLength(2);
    expect(grouped['pending']).toHaveLength(1);
  });
});

describe('totalBonusPaid', () => {
  it('sums bonuses for hired referrals', () => {
    expect(totalBonusPaid(referrals)).toBe(3500);
  });

  it('returns 0 when no hires', () => {
    expect(totalBonusPaid([makeReferral('pending')])).toBe(0);
  });
});

describe('referralHireRate', () => {
  it('calculates hire rate percentage', () => {
    expect(referralHireRate(referrals)).toBe(40);
  });

  it('returns 0 for empty array', () => {
    expect(referralHireRate([])).toBe(0);
  });
});

describe('topReferrers', () => {
  it('returns referrer IDs sorted by hires', () => {
    const top = topReferrers(referrals);
    expect(top[0]).toBe('emp1');
  });
});

describe('activeReferrals', () => {
  it('returns pending and interviewing referrals', () => {
    expect(activeReferrals(referrals)).toHaveLength(2);
  });
});
