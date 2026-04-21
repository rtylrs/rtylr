import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

// No proto type yet — define locally.
export interface StockTransfer {
  id: number;
  fromWarehouseId: number;
  toWarehouseId: number;
  fromName: string;
  toName: string;
  items: number;
  status: 'Draft' | 'In Transit' | 'Received' | 'Cancelled';
  createdAt: string;
  completedAt?: string;
}

export interface CreateStockTransferInput {
  fromWarehouseId: number;
  toWarehouseId: number;
  items: Array<{ productId: number; quantity: number }>;
  notes?: string;
}

interface ListStockTransfersResponse {
  stockTransfers: StockTransfer[];
}

export const useListStockTransfers = (): UseQueryResult<StockTransfer[]> => {
  return useQuery<StockTransfer[]>({
    queryKey: ['erp-stock-transfers'],
    queryFn: async (): Promise<StockTransfer[]> => {
      const res = await unwrap(safe.get<ListStockTransfersResponse>('/erp/stock-transfers'));
      return (res as ListStockTransfersResponse).stockTransfers ?? [];
    },
  });
};

export const useCreateStockTransfer = (): UseMutationResult<
  StockTransfer,
  Error,
  CreateStockTransferInput
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateStockTransferInput): Promise<StockTransfer> => {
      const res = await unwrap(safe.post<StockTransfer>('/erp/stock-transfers', data));
      return res as StockTransfer;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-stock-transfers'] });
      queryClient.invalidateQueries({ queryKey: ['erp-inventory-report'] });
      toast.success(t('erp.stockTransferCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
