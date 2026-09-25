import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { InventoryServiceClient } from '@voxire/voxie/erp/inventory/v1/client';

// ERP is a foreign domain here: this app reads its valuation read-only and
// never writes it. The client id stays 'finance' (the caller's app), matching
// how useDirectory reads HR from Flow.
const inventoryClient = createOneKitClient(InventoryServiceClient, 'finance');

export interface ErpValuation {
  /** Total current value of stock on hand, major units. */
  totalValue: number;
}

export const ERP_VALUATION_QUERY_KEY = ['erp', 'valuation'] as const;

/**
 * Reads ERP's inventory valuation for the balance-sheet Inventory line.
 *
 * Degrade behaviour (#845): the section renders only when this query
 * succeeds with data. A tenant without ERP (or a failed read) yields an
 * error state the caller must treat as "hide", never as zero: a balance
 * sheet that silently shows 0 inventory misstates the business.
 *
 * Entitlement note: there is no statically-known ERP addon id to gate on
 * (the catalog is DB-seeded), so the read itself is the probe. Swapping in
 * `enabled: hasErpAddon` becomes a one-line change if dash's addon ids gain
 * a shared constant.
 */
export function useErpInventoryValuation() {
  return useQuery<ErpValuation | null>({
    queryKey: ERP_VALUATION_QUERY_KEY,
    queryFn: async (): Promise<ErpValuation | null> => {
      try {
        const res = await inventoryClient.getValuation({});
        if (res === null || res === undefined) return null;
        const items = res.items ?? [];
        if (!items.length && !res.totalValue) return null;
        return { totalValue: Number(res.totalValue ?? 0) };
      } catch {
        // No ERP for this tenant, or the read failed: hide rather than zero.
        return null;
      }
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
