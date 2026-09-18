// Pure helpers for customer display and classification.

import type { Customer } from '../types/pos';

/** Return a display name with a fallback for customers without a name. */
export function displayName(customer: Customer | null | undefined, fallback = 'Guest'): string {
  if (!customer || !customer.name.trim()) return fallback;
  return customer.name.trim();
}

/** Return true if the customer has a loyalty points balance. */
export function hasLoyaltyPoints(customer: Customer): boolean {
  return typeof customer.loyaltyPoints === 'number' && customer.loyaltyPoints > 0;
}

/** Return the loyalty points balance, defaulting to 0. */
export function loyaltyBalance(customer: Customer): number {
  return typeof customer.loyaltyPoints === 'number' && Number.isFinite(customer.loyaltyPoints)
    ? Math.max(0, customer.loyaltyPoints)
    : 0;
}

/** Return true if the customer has a valid phone number set. */
export function hasPhone(customer: Customer): boolean {
  return typeof customer.phone === 'string' && customer.phone.trim().length > 0;
}

/** Return true if the customer has a valid email address set. */
export function hasEmail(customer: Customer): boolean {
  return typeof customer.email === 'string' && customer.email.includes('@');
}

/** Return the customer's first name (everything before the first space). */
export function firstName(customer: Customer): string {
  return customer.name.trim().split(' ')[0];
}

/** Return a short label suitable for a receipt or display (first name + last initial). */
export function shortName(customer: Customer): string {
  const parts = customer.name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0];
  return `${parts[0]} ${parts[parts.length - 1][0]}.`;
}

/** Return true if the customer has visited before (has a lastVisit date). */
export function isReturningCustomer(customer: Customer): boolean {
  return typeof customer.lastVisit === 'string' && customer.lastVisit.length > 0;
}

/** Sort customers alphabetically by name. */
export function sortByName(customers: Customer[]): Customer[] {
  return [...customers].sort((a, b) => a.name.localeCompare(b.name));
}
