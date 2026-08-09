import { describe, expect, it } from 'vitest';
import {
  annualPremiumCost,
  benefitStatusColor,
  benefitStatusLabel,
  employeePremiumShare,
  employerPremiumShare,
  filterByBenefitType,
  isBenefitActive,
  isBenefitExpired,
  isEnrollmentValid,
  totalActivePremium,
} from './benefitsUtils';

describe('benefitStatusLabel', () => {
  it('returns correct label for each status', () => {
    expect(benefitStatusLabel('Active')).toBe('Active');
    expect(benefitStatusLabel('Inactive')).toBe('Inactive');
    expect(benefitStatusLabel('Pending')).toBe('Pending');
    expect(benefitStatusLabel('Expired')).toBe('Expired');
  });
});

describe('benefitStatusColor', () => {
  it('returns green for Active', () => {
    expect(benefitStatusColor('Active')).toBe('green');
  });

  it('returns gray for Inactive', () => {
    expect(benefitStatusColor('Inactive')).toBe('gray');
  });

  it('returns yellow for Pending', () => {
    expect(benefitStatusColor('Pending')).toBe('yellow');
  });

  it('returns red for Expired', () => {
    expect(benefitStatusColor('Expired')).toBe('red');
  });
});

describe('isBenefitActive', () => {
  it('returns true only for Active', () => {
    expect(isBenefitActive('Active')).toBe(true);
    expect(isBenefitActive('Inactive')).toBe(false);
    expect(isBenefitActive('Pending')).toBe(false);
    expect(isBenefitActive('Expired')).toBe(false);
  });
});

describe('isBenefitExpired', () => {
  it('returns true only for Expired', () => {
    expect(isBenefitExpired('Expired')).toBe(true);
    expect(isBenefitExpired('Active')).toBe(false);
  });
});

describe('employeePremiumShare', () => {
  it('calculates employee share correctly', () => {
    expect(employeePremiumShare(100, 0.8)).toBeCloseTo(20);
    expect(employeePremiumShare(200, 0.5)).toBeCloseTo(100);
    expect(employeePremiumShare(100, 1)).toBeCloseTo(0);
    expect(employeePremiumShare(100, 0)).toBeCloseTo(100);
  });

  it('clamps rate to [0, 1]', () => {
    expect(employeePremiumShare(100, 1.5)).toBeCloseTo(0);
    expect(employeePremiumShare(100, -0.5)).toBeCloseTo(100);
  });
});

describe('employerPremiumShare', () => {
  it('calculates employer share correctly', () => {
    expect(employerPremiumShare(100, 0.8)).toBeCloseTo(80);
    expect(employerPremiumShare(200, 0.5)).toBeCloseTo(100);
  });
});

describe('annualPremiumCost', () => {
  it('multiplies monthly by 12', () => {
    expect(annualPremiumCost(100)).toBe(1200);
    expect(annualPremiumCost(0)).toBe(0);
    expect(annualPremiumCost(50.5)).toBeCloseTo(606);
  });
});

describe('isEnrollmentValid', () => {
  it('returns true when enrolled date is before or equal to effective date', () => {
    expect(isEnrollmentValid('2024-01-01', '2024-01-15')).toBe(true);
    expect(isEnrollmentValid('2024-01-15', '2024-01-15')).toBe(true);
  });

  it('returns false when enrolled date is after effective date', () => {
    expect(isEnrollmentValid('2024-02-01', '2024-01-15')).toBe(false);
  });
});

describe('filterByBenefitType', () => {
  const benefits = [
    { type: 'Health' as const, name: 'Medical' },
    { type: 'Dental' as const, name: 'Dental' },
    { type: 'Health' as const, name: 'Medical Plus' },
  ];

  it('filters by benefit type', () => {
    const result = filterByBenefitType(benefits, 'Health');
    expect(result).toHaveLength(2);
    expect(result.every((b) => b.type === 'Health')).toBe(true);
  });

  it('returns empty array when no match', () => {
    expect(filterByBenefitType(benefits, 'Vision')).toHaveLength(0);
  });
});

describe('totalActivePremium', () => {
  it('sums monthly premiums for active benefits only', () => {
    const benefits = [
      { status: 'Active' as const, monthlyPremium: 100 },
      { status: 'Inactive' as const, monthlyPremium: 50 },
      { status: 'Active' as const, monthlyPremium: 75 },
    ];
    expect(totalActivePremium(benefits)).toBe(175);
  });

  it('returns 0 when no active benefits', () => {
    const benefits = [{ status: 'Expired' as const, monthlyPremium: 100 }];
    expect(totalActivePremium(benefits)).toBe(0);
  });
});
