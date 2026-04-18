import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi, InventoryReportItem, LowStockItem } from '../../shared/services/erpApi';
import { inventoryReportItemSchema, lowStockItemSchema } from '../../shared/schemas/inventory';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useInventoryReport() {
  return useQuery({
    queryKey: ['inventory-report'],
    queryFn: async () => {
      const res = await unwrap(erpApi.getInventoryReport());
      return z
        .array(inventoryReportItemSchema)
        .parse(unwrapList(res, 'items', 'report')) as InventoryReportItem[];
    },
  });
}

export function useLowStock() {
  return useQuery({
    queryKey: ['low-stock'],
    queryFn: async () => {
      const res = await unwrap(erpApi.getLowStock());
      return z
        .array(lowStockItemSchema)
        .parse(unwrapList(res, 'items', 'low_stock')) as LowStockItem[];
    },
  });
}

export function useAdjustInventory() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { productId: number; delta: number }) =>
      unwrap(erpApi.adjustInventory(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory-report'] });
      qc.invalidateQueries({ queryKey: ['low-stock'] });
      qc.invalidateQueries({ queryKey: ['products'] });
    },
  });
}
