import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  InventoryServiceClient,
  type AdjustInventoryRequest,
  type AdjustInventoryResponse,
  type InventoryReportItem,
} from '@voxire/voxie/erp/inventory/v1/service_client';

const inventoryClient = createOneKitClient(InventoryServiceClient, 'erp');

export type { InventoryReportItem };

export type AdjustInventoryInput = AdjustInventoryRequest;

export function useInventoryReport() {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['inventory-report'],
    queryFn: async () => {
      const response = await inventoryClient.getInventoryReport({});
      return response.items;
    },
  });
}

export function useLowStock() {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['low-stock'],
    queryFn: async () => {
      const response = await inventoryClient.getLowStock({});
      return response.items;
    },
  });
}

export function useAdjustInventory() {
  const qc = useQueryClient();
  return useMutation<AdjustInventoryResponse, Error, AdjustInventoryInput>({
    mutationFn: (data) => inventoryClient.adjustInventory(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['inventory-report'] });
      qc.invalidateQueries({ queryKey: ['low-stock'] });
      qc.invalidateQueries({ queryKey: ['products'] });
    },
    onError: () => toast.error('Failed to adjust inventory'),
  });
}
