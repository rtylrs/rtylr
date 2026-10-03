import type { QueryClient } from '@tanstack/react-query';

export const STOCK_DERIVED_QUERY_KEYS = [
  ['inventory-report'],
  ['low-stock'],
  ['inventory-valuation'],
  ['reorder-suggestions'],
  ['expiry'],
  ['products'],
  ['product'],
  ['product-full'],
  ['erp-dashboard-stats'],
  ['pos-menu-items'],
  ['menus'],
  ['menu'],
] as const;

export function invalidateStockDerivedQueries(
  queryClient: Pick<QueryClient, 'invalidateQueries'>
): void {
  for (const queryKey of STOCK_DERIVED_QUERY_KEYS) {
    void queryClient.invalidateQueries({ queryKey });
  }
}
