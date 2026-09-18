import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { InventoryServiceClient } from '@voxire/voxie/erp/inventory/v1/client';
import type {
  AdjustInventoryRequest,
  AdjustInventoryResponse,
  InventoryReportItem,
} from '@voxire/voxie/erp/inventory/v1/types';
import { withErpListBounds } from '../../shared/lib/erpListBounds';
import { invalidateStockDerivedQueries } from '../../shared/lib/stockInvalidation';

const inventoryClient = createOneKitClient(InventoryServiceClient, 'erp');

export type { InventoryReportItem };

export type AdjustInventoryInput = AdjustInventoryRequest;

export function useInventoryReport() {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['inventory-report'],
    queryFn: async () => {
      const response = await inventoryClient.getInventoryReport(withErpListBounds({}));
      return response.items ?? [];
    },
  });
}

export function useLowStock() {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['low-stock'],
    queryFn: async () => {
      const response = await inventoryClient.getLowStock(withErpListBounds({}));
      return response.items ?? [];
    },
  });
}

export function useAdjustInventory() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<AdjustInventoryResponse, Error, AdjustInventoryInput>({
    mutationFn: (data) => inventoryClient.adjustInventory(data),
    onSuccess: () => {
      invalidateStockDerivedQueries(qc);
    },
    onError: () => toast.error(t('erp.inventory.adjustFailed')),
  });
}
