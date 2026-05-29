/**
 * Employee benefits helpers for the HR application.
 * All functions are pure with no side-effects.
 */

export type BenefitStatus = 'Active' | 'Inactive' | 'Pending' | 'Expired';
export type BenefitType = 'Health' | 'Dental' | 'Vision' | 'Life' | 'Disability' | string;

/** Returns a display label for a benefit status. */
export function benefitStatusLabel(status: BenefitStatus): string {
  return status;
}

/** Returns a badge color for a benefit status. */
export function benefitStatusColor(status: BenefitStatus): string {
  switch (status) {
    case 'Active':
      return 'green';
    case 'Inactive':
      return 'gray';
    case 'Pending':
      return 'yellow';
    case 'Expired':
      return 'red';
    default:
      return 'gray';
  }
}

/** Returns true if the benefit status is Active. */
export function isBenefitActive(status: BenefitStatus): boolean {
  return status === 'Active';
}

/** Returns true if the benefit status is Expired. */
export function isBenefitExpired(status: BenefitStatus): boolean {
  return status === 'Expired';
}

/**
 * Calculates the employee's monthly premium share.
 * @param totalPremium - Total monthly premium amount.
 * @param employerRate - Fraction of premium the employer covers (0 to 1).
 */
export function employeePremiumShare(totalPremium: number, employerRate: number): number {
  const clamped = Math.min(1, Math.max(0, employerRate));
  return totalPremium * (1 - clamped);
}

/**
 * Calculates the employer's monthly premium share.
 * @param totalPremium - Total monthly premium amount.
 * @param employerRate - Fraction of premium the employer covers (0 to 1).
 */
export function employerPremiumShare(totalPremium: number, employerRate: number): number {
  const clamped = Math.min(1, Math.max(0, employerRate));
  return totalPremium * clamped;
}

/** Returns the annual cost of a benefit from its monthly premium. */
export function annualPremiumCost(monthlyPremium: number): number {
  return monthlyPremium * 12;
}

/**
 * Returns true if the enrollment date is on or before the effective date.
 * Both parameters are ISO date strings (YYYY-MM-DD).
 */
export function isEnrollmentValid(enrolledDate: string, effectiveDate: string): boolean {
  return enrolledDate <= effectiveDate;
}

/**
 * Filters a list of benefit records by type.
 */
export function filterByBenefitType<T extends { type: BenefitType }>(
  benefits: T[],
  type: BenefitType
): T[] {
  return benefits.filter((b) => b.type === type);
}

/**
 * Sums the monthly premiums for all active benefits in a list.
 */
export function totalActivePremium<T extends { status: BenefitStatus; monthlyPremium: number }>(
  benefits: T[]
): number {
  return benefits
    .filter((b) => isBenefitActive(b.status))
    .reduce((acc, b) => acc + b.monthlyPremium, 0);
}
