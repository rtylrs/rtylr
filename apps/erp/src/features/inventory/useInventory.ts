import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { InventoryServiceClient } from '@voxire/voxie/erp/inventory/v1/client';
import type {
  AdjustInventoryRequest,
  AdjustInventoryResponse,
  InventoryReportItem,
} from '@voxire/voxie/erp/inventory/v1/types';
import { invalidateStockDerivedQueries } from '../../shared/lib/stockInvalidation';

const inventoryClient = createOneKitClient(InventoryServiceClient, 'erp');

export type { InventoryReportItem };

export type AdjustInventoryInput = AdjustInventoryRequest;

export function useInventoryReport() {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['inventory-report'],
    queryFn: async () => {
      const response = await inventoryClient.getInventoryReport({});
      return response.items ?? [];
    },
  });
}

export function useLowStock() {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['low-stock'],
    queryFn: async () => {
      const response = await inventoryClient.getLowStock({});
      return response.items ?? [];
    },
  });
}

export function useAdjustInventory() {
  const qc = useQueryClient();
  return useMutation<AdjustInventoryResponse, Error, AdjustInventoryInput>({
    mutationFn: (data) => inventoryClient.adjustInventory(data),
    onSuccess: () => {
      invalidateStockDerivedQueries(qc);
    },
    onError: () => toast.error('Failed to adjust inventory'),
  });
}
