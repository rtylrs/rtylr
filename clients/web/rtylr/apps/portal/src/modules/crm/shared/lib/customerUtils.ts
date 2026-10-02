import type { Customer } from '../schemas/customers';

/** Returns true if the customer has opted in to email marketing. */
export function hasEmailConsent(customer: Customer): boolean {
  return customer.consent?.email === true;
}

/** Returns true if the customer has opted in to SMS marketing. */
export function hasSmsConsent(customer: Customer): boolean {
  return customer.consent?.sms === true;
}

/** Returns the number of days since the customer's last visit. */
export function daysSinceLastVisit(customer: Customer, now = new Date()): number {
  const last = new Date(customer.lastVisitAt ?? customer.lastVisit);
  if (!Number.isFinite(last.getTime())) return Number.POSITIVE_INFINITY;
  const diff = now.getTime() - last.getTime();
  return Math.floor(diff / (1000 * 60 * 60 * 24));
}

/** Returns true if the customer is at risk (no visit in atRiskDays). */
export function isAtRisk(customer: Customer, atRiskDays: number, now = new Date()): boolean {
  return daysSinceLastVisit(customer, now) >= atRiskDays;
}

/** Returns the customer's tags or an empty array. */
export function getTags(customer: Customer): string[] {
  return customer.tags ?? [];
}

/** Returns true if the customer has a given tag. */
export function hasTag(customer: Customer, tag: string): boolean {
  return getTags(customer).includes(tag);
}

/** Returns the customer's display label (name + email). */
export function customerLabel(customer: Customer): string {
  return `${customer.name} <${customer.email}>`;
}
