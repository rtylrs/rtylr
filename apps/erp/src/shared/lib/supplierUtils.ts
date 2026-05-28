import type { Supplier } from '../services/erpApi';

/** Sorts suppliers alphabetically by name, ascending. Does not mutate. */
export function sortSuppliersByName(suppliers: Supplier[]): Supplier[] {
  return [...suppliers].sort((a, b) => a.name.localeCompare(b.name));
}

/** Sorts suppliers by product count descending (highest first). Does not mutate. */
export function sortSuppliersByProductCount(suppliers: Supplier[]): Supplier[] {
  return [...suppliers].sort((a, b) => b.productCount - a.productCount);
}

/**
 * Sorts suppliers by on-time delivery percentage descending.
 * Suppliers without OTD data are placed at the end.
 * Does not mutate.
 */
export function sortSuppliersByOnTimeDelivery(suppliers: Supplier[]): Supplier[] {
  return [...suppliers].sort((a, b) => {
    const aVal = a.onTimeDeliveryPct ?? -1;
    const bVal = b.onTimeDeliveryPct ?? -1;
    return bVal - aVal;
  });
}

/** Returns only suppliers with status 'Active'. */
export function filterActiveSuppliers(suppliers: Supplier[]): Supplier[] {
  return suppliers.filter((s) => s.status === 'Active');
}

/** Returns suppliers that carry the given category. */
export function filterSuppliersByCategory(suppliers: Supplier[], category: string): Supplier[] {
  return suppliers.filter((s) => s.categories.includes(category));
}

/**
 * Computes a weighted performance score for a supplier:
 * - OTD (on-time delivery) counts 60%
 * - Fulfillment accuracy counts 40%
 * Returns null when neither metric is available.
 */
export function supplierPerformanceScore(supplier: Supplier): number | null {
  const { onTimeDeliveryPct: otd, fulfillmentAccuracy: fa } = supplier;
  if (otd === undefined && fa === undefined) return null;
  if (otd !== undefined && fa !== undefined) {
    return Math.round(otd * 0.6 + fa * 0.4);
  }
  return otd ?? fa ?? null;
}

/**
 * Returns the first-letter initials of up to two words in the supplier name,
 * uppercased.
 */
export function supplierInitials(name: string): string {
  return name
    .split(/\s+/)
    .slice(0, 2)
    .map((w) => w[0].toUpperCase())
    .join('');
}
