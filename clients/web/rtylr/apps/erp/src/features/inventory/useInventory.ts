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
import { ERP_LIST_LIMIT } from '../../shared/lib/erpListBounds';
import { invalidateStockDerivedQueries } from '../../shared/lib/stockInvalidation';

const inventoryClient = createOneKitClient(InventoryServiceClient, 'erp');

export type { InventoryReportItem };

export type AdjustInventoryInput = AdjustInventoryRequest;

type InventoryPage<T> = {
  items?: T[];
  hasMore?: boolean;
  total?: number;
};

async function fetchAllPages<T>(fetchPage: (offset: number) => Promise<InventoryPage<T>>) {
  const items: T[] = [];
  let offset = 0;

  while (true) {
    const response = await fetchPage(offset);
    const page = response.items ?? [];
    items.push(...page);

    const hasMore =
      response.hasMore ??
      (response.total !== undefined
        ? items.length < response.total
        : page.length === ERP_LIST_LIMIT);
    if (!hasMore || page.length === 0) return items;
    offset += page.length;
  }
}

export function fetchAllInventoryReport(): Promise<InventoryReportItem[]> {
  return fetchAllPages((offset) =>
    inventoryClient.getInventoryReport({ limit: ERP_LIST_LIMIT, offset })
  );
}

export function fetchAllLowStock(): Promise<InventoryReportItem[]> {
  return fetchAllPages((offset) => inventoryClient.getLowStock({ limit: ERP_LIST_LIMIT, offset }));
}

export function useInventoryReport() {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['inventory-report'],
    queryFn: fetchAllInventoryReport,
  });
}

export function useLowStock() {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['low-stock'],
    queryFn: fetchAllLowStock,
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
