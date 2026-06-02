import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';
import { unwrapList } from '../../shared/lib/unwrapList';
import {
  stockTransferSchema,
  StockTransfer,
  CreateStockTransferInput,
  UpdateStockTransferStatusInput,
} from './stockTransferSchema';

export interface StockTransferListResult {
  transfers: StockTransfer[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export interface UseStockTransfersParams {
  page?: number;
  pageSize?: number;
}

export function useStockTransfers(params: UseStockTransfersParams = {}) {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;
  return useQuery<StockTransferListResult>({
    queryKey: ['stock-transfers', { page, pageSize }],
    queryFn: async () => {
      const res = await unwrap(
        safe.get<unknown>('/erp/stock-transfers', { params: { page, page_size: pageSize } })
      );
      const transfers = z
        .array(stockTransferSchema)
        .parse(unwrapList(res, 'stockTransfers', 'stock_transfers'));
      const hasMore = (res as Record<string, unknown>)?.hasMore === true;
      return { transfers, hasMore, page, pageSize };
    },
  });
}

export function useCreateStockTransfer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateStockTransferInput) =>
      unwrap(safe.post<unknown>('/erp/stock-transfers', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-transfers'] }),
    onError: () => toast.error('Failed to create stock transfer'),
  });
}

export function useUpdateStockTransferStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateStockTransferStatusInput) =>
      unwrap(safe.put<unknown>('/erp/stock-transfers', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-transfers'] }),
    onError: () => toast.error('Failed to update stock transfer'),
  });
}
