// Headcount planning helpers for the recruit admin panel.

export interface HeadcountPlan {
  department: string;
  current: number;
  target: number;
}

/**
 * headcountGap returns the difference between target and current headcount.
 */
export function headcountGap(plan: HeadcountPlan): number {
  return plan.target - plan.current;
}

/**
 * isAtCapacity returns true when current headcount meets or exceeds target.
 */
export function isAtCapacity(plan: HeadcountPlan): boolean {
  return plan.current >= plan.target;
}

/**
 * totalCurrentHeadcount sums current headcount across all plans.
 */
export function totalCurrentHeadcount(plans: HeadcountPlan[]): number {
  return plans.reduce((sum, p) => sum + p.current, 0);
}

/**
 * totalTargetHeadcount sums target headcount across all plans.
 */
export function totalTargetHeadcount(plans: HeadcountPlan[]): number {
  return plans.reduce((sum, p) => sum + p.target, 0);
}

/**
 * openRoles returns the total number of open roles across all plans.
 */
export function openRoles(plans: HeadcountPlan[]): number {
  return plans.reduce((sum, p) => sum + Math.max(0, headcountGap(p)), 0);
}

/**
 * departmentsUnderTarget returns plans where current is below target.
 */
export function departmentsUnderTarget(plans: HeadcountPlan[]): HeadcountPlan[] {
  return plans.filter((p) => !isAtCapacity(p));
}

/**
 * hiringProgress returns the percentage of target filled across all plans.
 */
export function hiringProgress(plans: HeadcountPlan[]): number {
  const target = totalTargetHeadcount(plans);
  if (target === 0) return 0;
  return Math.round((totalCurrentHeadcount(plans) / target) * 1000) / 10;
}
