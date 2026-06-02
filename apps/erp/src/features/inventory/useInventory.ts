import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi } from '../../shared/services/erpApi';
import { safe } from '../../shared/lib/api';
import { inventoryReportItemSchema, type InventoryReportItem } from './inventorySchemas';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useInventoryReport() {
  return useQuery({
    queryKey: ['inventory-report'],
    queryFn: async (): Promise<InventoryReportItem[]> => {
      const res = await unwrap(erpApi.getInventoryReport());
      return z.array(inventoryReportItemSchema).parse(unwrapList(res, 'items', 'report'));
    },
  });
}

export function useLowStock() {
  return useQuery({
    queryKey: ['low-stock'],
    queryFn: async (): Promise<InventoryReportItem[]> => {
      const res = await unwrap(erpApi.getLowStock());
      // Backend returns the same InventoryReportItem row shape under `items`.
      return z.array(inventoryReportItemSchema).parse(unwrapList(res, 'items', 'low_stock'));
    },
  });
}

/**
 * Backend POST /erp/inventory/adjust expects { productId, change, type, notes }.
 * `type` must be one of: purchase, sale, adjustment, return, waste.
 */
export interface AdjustInventoryInput {
  productId: number;
  change: number;
  type: 'purchase' | 'sale' | 'adjustment' | 'return' | 'waste';
  notes?: string;
}

export function useAdjustInventory() {
  const qc = useQueryClient();
  return useMutation({
    // Call the endpoint directly with the body the backend handler expects.
    // The shared erpApi.adjustInventory helper still types the body as
    // { productId, delta } which does not match the backend contract.
    mutationFn: (data: AdjustInventoryInput) =>
      unwrap(
        safe.post<{ message: string }>('/erp/inventory/adjust', {
          productId: data.productId,
          change: data.change,
          type: data.type,
          notes: data.notes ?? '',
        })
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory-report'] });
      qc.invalidateQueries({ queryKey: ['low-stock'] });
      qc.invalidateQueries({ queryKey: ['products'] });
    },
    onError: () => toast.error('Failed to adjust inventory'),
  });
}
