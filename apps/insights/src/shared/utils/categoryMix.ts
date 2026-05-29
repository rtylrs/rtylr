// Product category mix helpers for the insights products feature.

import type { CategorySales } from '../hooks/queries/types';

/**
 * totalCategorySales sums revenueMinorUnits across all categories.
 */
export function totalCategorySales(categories: CategorySales[]): number {
  return categories.reduce((sum, c) => sum + c.revenueMinorUnits, 0);
}

/**
 * topCategories returns the n categories with the highest revenue.
 */
export function topCategories(categories: CategorySales[], n: number): CategorySales[] {
  return [...categories].sort((a, b) => b.revenueMinorUnits - a.revenueMinorUnits).slice(0, n);
}

/**
 * categoryShare returns the percentage of total revenue for each category.
 * Returns the input array with a computed pct field (overwrites existing pct).
 */
export function categoryShare(
  categories: CategorySales[]
): (CategorySales & { computedPct: number })[] {
  const total = totalCategorySales(categories);
  if (total === 0) return categories.map((c) => ({ ...c, computedPct: 0 }));
  return categories.map((c) => ({
    ...c,
    computedPct: Math.round((c.revenueMinorUnits / total) * 10000) / 100,
  }));
}

/**
 * categoryLabels extracts the category name from each entry.
 */
export function categoryLabels(categories: CategorySales[]): string[] {
  return categories.map((c) => c.categoryName);
}

/**
 * categoryValues extracts the revenueMinorUnits from each entry.
 */
export function categoryValues(categories: CategorySales[]): number[] {
  return categories.map((c) => c.revenueMinorUnits);
}
