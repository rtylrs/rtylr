// Product analytics helpers for the insights products feature.

import type { BestSeller, MarginBucket, SlowMover } from '../hooks/queries/types';

/**
 * topSellersByRevenue returns the n best-selling products sorted by
 * revenueMinorUnits descending.
 */
export function topSellersByRevenue(products: BestSeller[], n: number): BestSeller[] {
  return [...products].sort((a, b) => b.revenueMinorUnits - a.revenueMinorUnits).slice(0, n);
}

/**
 * topSellersByUnits returns the n best-selling products sorted by
 * unitsSold descending.
 */
export function topSellersByUnits(products: BestSeller[], n: number): BestSeller[] {
  return [...products].sort((a, b) => b.unitsSold - a.unitsSold).slice(0, n);
}

/**
 * avgRevenuePerUnit returns the average revenue per unit sold for a
 * product. Returns 0 when unitsSold is 0.
 */
export function avgRevenuePerUnit(product: BestSeller): number {
  if (product.unitsSold === 0) return 0;
  return product.revenueMinorUnits / product.unitsSold;
}

/**
 * totalProductsInMarginBuckets sums productCount across all margin buckets.
 */
export function totalProductsInMarginBuckets(buckets: MarginBucket[]): number {
  return buckets.reduce((sum, b) => sum + b.productCount, 0);
}

/**
 * highMarginBuckets returns buckets where lowerPct meets or exceeds
 * the given threshold.
 */
export function highMarginBuckets(buckets: MarginBucket[], thresholdPct: number): MarginBucket[] {
  return buckets.filter((b) => b.lowerPct >= thresholdPct);
}

/**
 * slowMoversByStock returns slow movers sorted by stockQuantity
 * descending - highest idle inventory first.
 */
export function slowMoversByStock(products: SlowMover[]): SlowMover[] {
  return [...products].sort((a, b) => b.stockQuantity - a.stockQuantity);
}

/**
 * totalSlowMoverStock sums the stockQuantity of all slow movers.
 */
export function totalSlowMoverStock(products: SlowMover[]): number {
  return products.reduce((sum, p) => sum + p.stockQuantity, 0);
}
