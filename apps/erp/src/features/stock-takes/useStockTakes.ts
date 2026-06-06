import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../shared/lib/api';
import { unwrapList } from '../../shared/lib/unwrapList';
import {
  stockTakeSchema,
  StockTake,
  CreateStockTakeInput,
  UpdateStockTakeInput,
} from './stockTakeSchema';

export interface StockTakeListResult {
  stockTakes: StockTake[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export interface UseStockTakesParams {
  page?: number;
  pageSize?: number;
}

export function useStockTakes(params: UseStockTakesParams = {}) {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;
  return useQuery<StockTakeListResult>({
    queryKey: ['stock-takes', { page, pageSize }],
    queryFn: async () => {
      const res = await unwrap(
        safe.get<unknown>('/erp/stock-takes', { params: { page, page_size: pageSize } })
      );
      const stockTakes = z
        .array(stockTakeSchema)
        .parse(unwrapList(res, 'stockTakes', 'stock_takes'));
      const hasMore = (res as Record<string, unknown>)?.hasMore === true;
      return { stockTakes, hasMore, page, pageSize };
    },
  });
}

export function useCreateStockTake() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateStockTakeInput) =>
      unwrap(safe.post<unknown>('/erp/stock-takes', data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-takes'] }),
    onError: () => toast.error('Failed to create stock take'),
  });
}

export function useUpdateStockTake() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateStockTakeInput) =>
      unwrap(safe.put<unknown>(`/erp/stock-takes/${data.id}`, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-takes'] }),
    onError: () => toast.error('Failed to update stock take'),
  });
}
