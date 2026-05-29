/**
 * Product status and lifecycle helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

import type { Product } from '../services/erpApi';

export type StockStatus = 'in-stock' | 'low-stock' | 'out-of-stock';
export type ProfitabilityBand = 'high' | 'medium' | 'low' | 'loss';

/** Compute the stock status from current stock and low-stock threshold. */
export function stockStatus(stock: number, lowStock: number): StockStatus {
  if (stock === 0) return 'out-of-stock';
  if (stock <= lowStock) return 'low-stock';
  return 'in-stock';
}

/** Return a user-facing label for a stock status. */
export function stockStatusLabel(status: StockStatus): string {
  switch (status) {
    case 'in-stock':
      return 'In Stock';
    case 'low-stock':
      return 'Low Stock';
    case 'out-of-stock':
      return 'Out of Stock';
  }
}

/** Return the UI color for a stock status badge. */
export function stockStatusColor(status: StockStatus): string {
  switch (status) {
    case 'in-stock':
      return 'green';
    case 'low-stock':
      return 'yellow';
    case 'out-of-stock':
      return 'red';
  }
}

/** Compute gross margin percentage (0-100) for a product. */
export function grossMarginPct(product: Product): number {
  if (product.salePrice === 0) return 0;
  return Math.round(((product.salePrice - product.costPrice) / product.salePrice) * 100);
}

/** Classify a product into a profitability band. */
export function profitabilityBand(product: Product): ProfitabilityBand {
  const margin = grossMarginPct(product);
  if (margin >= 40) return 'high';
  if (margin >= 20) return 'medium';
  if (margin >= 0) return 'low';
  return 'loss';
}

/** Filter products by stock status. */
export function filterByStockStatus(products: Product[], status: StockStatus): Product[] {
  return products.filter((p) => stockStatus(p.stock, p.lowStock) === status);
}

/** Return products that need reordering (stock at or below reorder point). */
export function needsReorder(products: Product[]): Product[] {
  return products.filter((p) => {
    const threshold = p.reorderPoint ?? p.lowStock;
    return p.stock <= threshold;
  });
}

/** Compute the total inventory value for a list of products (cost-price based). */
export function totalInventoryValue(products: Product[]): number {
  return products.reduce((sum, p) => sum + p.costPrice * p.stock, 0);
}

/** Compute the potential revenue if all stock is sold at sale price. */
export function potentialRevenue(products: Product[]): number {
  return products.reduce((sum, p) => sum + p.salePrice * p.stock, 0);
}

/** Return the top N products by inventory value. */
export function topByValue(products: Product[], n = 10): Product[] {
  return [...products].sort((a, b) => b.costPrice * b.stock - a.costPrice * a.stock).slice(0, n);
}
