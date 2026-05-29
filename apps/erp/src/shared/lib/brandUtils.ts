/**
 * Brand helper functions used across the ERP application.
 */
import type { Brand } from '../services/erpApi';

/** Sort brands alphabetically by name. */
export function sortBrandsByName(brands: Brand[]): Brand[] {
  return [...brands].sort((a, b) => a.name.localeCompare(b.name));
}

/** Sort brands by product count, highest first. */
export function sortBrandsByProductCount(brands: Brand[]): Brand[] {
  return [...brands].sort((a, b) => b.productCount - a.productCount);
}

/** Return the total product count across all brands. */
export function totalBrandProductCount(brands: Brand[]): number {
  return brands.reduce((sum, b) => sum + b.productCount, 0);
}

/** Return true when a brand has at least one product. */
export function hasProducts(brand: Brand): boolean {
  return brand.productCount > 0;
}

/** Return brand initials for avatar display (up to 2 characters). */
export function brandInitials(name: string): string {
  const words = name.trim().split(/\s+/);
  if (words.length === 1) return words[0].substring(0, 2).toUpperCase();
  return (words[0][0] + words[1][0]).toUpperCase();
}

/** Return brands that have at least one product. */
export function filterBrandsWithProducts(brands: Brand[]): Brand[] {
  return brands.filter(hasProducts);
}
