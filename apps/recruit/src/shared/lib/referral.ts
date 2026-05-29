// Employee referral program helpers for the recruit admin panel.

export interface Referral {
  id: string;
  referrerId: string;
  candidateName: string;
  position: string;
  status: 'pending' | 'interviewing' | 'hired' | 'rejected';
  submittedAt: string;
  bonusAmount?: number;
}

/**
 * referralsByStatus groups referrals by their current status.
 */
export function referralsByStatus(referrals: Referral[]): Record<string, Referral[]> {
  return referrals.reduce<Record<string, Referral[]>>((acc, r) => {
    acc[r.status] = [...(acc[r.status] ?? []), r];
    return acc;
  }, {});
}

/**
 * totalBonusPaid returns the sum of bonuses for hired referrals.
 */
export function totalBonusPaid(referrals: Referral[]): number {
  return referrals
    .filter((r) => r.status === 'hired' && r.bonusAmount)
    .reduce((sum, r) => sum + (r.bonusAmount ?? 0), 0);
}

/**
 * referralHireRate returns the percentage of referrals that resulted in a hire.
 */
export function referralHireRate(referrals: Referral[]): number {
  if (referrals.length === 0) return 0;
  const hired = referrals.filter((r) => r.status === 'hired').length;
  return Math.round((hired / referrals.length) * 1000) / 10;
}

/**
 * topReferrers returns referrer IDs sorted by number of hires.
 */
export function topReferrers(referrals: Referral[]): string[] {
  const hiresByReferrer = referrals
    .filter((r) => r.status === 'hired')
    .reduce<Record<string, number>>((acc, r) => {
      acc[r.referrerId] = (acc[r.referrerId] ?? 0) + 1;
      return acc;
    }, {});
  return Object.entries(hiresByReferrer)
    .sort(([, a], [, b]) => b - a)
    .map(([id]) => id);
}

/**
 * activeReferrals returns referrals that are pending or interviewing.
 */
export function activeReferrals(referrals: Referral[]): Referral[] {
  return referrals.filter((r) => r.status === 'pending' || r.status === 'interviewing');
}
