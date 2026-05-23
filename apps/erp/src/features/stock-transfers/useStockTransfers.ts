import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, StockTransfer } from '../../shared/services/erpApi';
import { stockTransferSchema } from '../../shared/schemas/stockTransfers';
import { unwrapList } from '../../shared/lib/unwrapList';

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
      const res = await unwrap(erpApi.listStockTransfers({ page, page_size: pageSize }));
      const transfers = z
        .array(stockTransferSchema)
        .parse(unwrapList(res, 'stock_transfers', 'stockTransfers')) as StockTransfer[];
      const hasMore = (res as unknown as Record<string, unknown>)?.has_more === true;
      return { transfers, hasMore, page, pageSize };
    },
  });
}

export function useCreateStockTransfer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<StockTransfer>) => unwrap(erpApi.createStockTransfer(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-transfers'] }),
    onError: () => toast.error('Failed to create stock transfer'),
  });
}
