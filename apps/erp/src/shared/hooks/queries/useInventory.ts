import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  AdjustInventoryRequest,
  AdjustInventoryResponse,
} from '@voxire/contracts/erp/inventory/v1/adjust_inventory';
import type { GetInventoryReportResponse } from '@voxire/contracts/erp/inventory/v1/get_inventory_report';
import type { GetLowStockResponse } from '@voxire/contracts/erp/inventory/v1/get_low_stock';
import type { GetValuationResponse } from '@voxire/contracts/erp/inventory/v1/get_valuation';
import type { InventoryReportItem, ValuationItem } from '@voxire/contracts/erp/inventory/v1/types';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface InventoryReportParams {
  warehouseId?: number;
}

interface ValuationParams {
  method: 'FIFO' | 'LIFO' | 'WAC';
}

export const useInventoryReport = (
  params: InventoryReportParams = {}
): UseQueryResult<GetInventoryReportResponse> => {
  return useQuery<GetInventoryReportResponse>({
    queryKey: ['erp-inventory-report', 'filter', params],
    queryFn: async (): Promise<GetInventoryReportResponse> => {
      const res = await unwrap(
        safe.get<GetInventoryReportResponse>('/erp/inventory/report', { params })
      );
      return res as GetInventoryReportResponse;
    },
  });
};

export const useLowStock = (): UseQueryResult<InventoryReportItem[]> => {
  return useQuery<InventoryReportItem[]>({
    queryKey: ['erp-low-stock'],
    queryFn: async (): Promise<InventoryReportItem[]> => {
      const res = await unwrap(safe.get<GetLowStockResponse>('/erp/inventory/low-stock'));
      return (res as GetLowStockResponse).items ?? [];
    },
  });
};

export const useInventoryValuation = (params: ValuationParams): UseQueryResult<ValuationItem[]> => {
  return useQuery<ValuationItem[]>({
    queryKey: ['erp-inventory-valuation', 'filter', params],
    queryFn: async (): Promise<ValuationItem[]> => {
      const res = await unwrap(
        safe.get<GetValuationResponse>('/erp/inventory/valuation', { params })
      );
      return (res as GetValuationResponse).items ?? [];
    },
  });
};

export const useAdjustInventory = (): UseMutationResult<
  AdjustInventoryResponse,
  Error,
  AdjustInventoryRequest
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: AdjustInventoryRequest): Promise<AdjustInventoryResponse> => {
      const res = await unwrap(safe.post<AdjustInventoryResponse>('/erp/inventory/adjust', data));
      return res as AdjustInventoryResponse;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-inventory-report'] });
      queryClient.invalidateQueries({ queryKey: ['erp-low-stock'] });
      queryClient.invalidateQueries({ queryKey: ['erp-inventory-valuation'] });
      queryClient.invalidateQueries({ queryKey: ['erp-products'] });
      toast.success(t('erp.inventoryAdjusted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
