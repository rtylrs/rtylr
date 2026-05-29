/**
 * Supplier performance scoring helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

import type { Supplier } from '../services/erpApi';

export type PerformanceTier = 'excellent' | 'good' | 'fair' | 'poor';

export interface SupplierScore {
  supplierId: number;
  name: string;
  score: number;
  tier: PerformanceTier;
}

/**
 * Compute a composite performance score (0-100) for a supplier.
 * Weights: on-time delivery 40%, fulfillment accuracy 40%, lead time 20%.
 */
export function computeSupplierScore(supplier: Supplier): number {
  const onTime = supplier.onTimeDeliveryPct ?? 0;
  const accuracy = supplier.fulfillmentAccuracy ?? 0;
  const leadTimeDays = supplier.leadTimeDays ?? 30;
  // Lead time score: 0 days = 100, 30+ days = 0
  const leadScore = Math.max(0, 100 - (leadTimeDays / 30) * 100);
  return Math.round(onTime * 0.4 + accuracy * 0.4 + leadScore * 0.2);
}

/** Map a numeric score to a performance tier label. */
export function scoreTier(score: number): PerformanceTier {
  if (score >= 85) return 'excellent';
  if (score >= 70) return 'good';
  if (score >= 50) return 'fair';
  return 'poor';
}

/** Build a scored and ranked list of suppliers. */
export function rankSuppliers(suppliers: Supplier[]): SupplierScore[] {
  return suppliers
    .filter((s) => s.status === 'Active')
    .map((s) => {
      const score = computeSupplierScore(s);
      return { supplierId: s.id, name: s.name, score, tier: scoreTier(score) };
    })
    .sort((a, b) => b.score - a.score);
}

/** Return suppliers below a given score threshold. */
export function underperformingSuppliers(suppliers: Supplier[], threshold = 70): SupplierScore[] {
  return rankSuppliers(suppliers).filter((s) => s.score < threshold);
}

/** Return the color associated with a performance tier for UI badges. */
export function tierColor(tier: PerformanceTier): string {
  switch (tier) {
    case 'excellent':
      return 'green';
    case 'good':
      return 'blue';
    case 'fair':
      return 'yellow';
    case 'poor':
      return 'red';
  }
}

/** Return the average score across all active suppliers. */
export function averageSupplierScore(suppliers: Supplier[]): number {
  const scored = rankSuppliers(suppliers);
  if (scored.length === 0) return 0;
  const total = scored.reduce((sum, s) => sum + s.score, 0);
  return Math.round(total / scored.length);
}
