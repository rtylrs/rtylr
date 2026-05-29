import type { Plan, BillingCycle } from '../components/pricing/pricingData';

/** Returns the display price string for a plan and billing cycle. */
export function getPlanPrice(plan: Plan, cycle: BillingCycle): string {
  return plan.price[cycle];
}

/** Returns the price note for a plan and billing cycle. */
export function getPlanPriceNote(plan: Plan, cycle: BillingCycle): string {
  return plan.priceNote[cycle];
}

/** Returns true if the plan is the featured / recommended plan. */
export function isFeaturedPlan(plan: Plan): boolean {
  return plan.featured === true;
}

/** Returns the featured plan from a list, or undefined if none is featured. */
export function getFeaturedPlan(plans: Plan[]): Plan | undefined {
  return plans.find(isFeaturedPlan);
}

/** Returns the savings percentage when switching from monthly to annual billing. */
export function annualSavingsPct(plan: Plan): number | null {
  const monthly = parseFloat(plan.price.monthly.replace(/[^0-9.]/g, ''));
  const annual = parseFloat(plan.price.annual.replace(/[^0-9.]/g, ''));
  if (isNaN(monthly) || isNaN(annual) || monthly === 0) return null;
  return Math.round(((monthly - annual) / monthly) * 100);
}

/** Returns plans that have a CTA action matching the given type. */
export function filterPlansByAction(plans: Plan[], action: Plan['ctaAction']): Plan[] {
  return plans.filter((p) => p.ctaAction === action);
}

/** Returns true if a plan has a free price for the given cycle. */
export function isFreePlan(plan: Plan, cycle: BillingCycle): boolean {
  const price = getPlanPrice(plan, cycle);
  return price === '$0' || price.toLowerCase() === 'free';
}
