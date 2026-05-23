import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, StockTake } from '../../shared/services/erpApi';
import { stockTakeSchema } from '../../shared/schemas/stockTakes';
import { unwrapList } from '../../shared/lib/unwrapList';

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
      const res = await unwrap(erpApi.listStockTakes({ page, page_size: pageSize }));
      const stockTakes = z
        .array(stockTakeSchema)
        .parse(unwrapList(res, 'stock_takes', 'stockTakes')) as StockTake[];
      const hasMore = (res as unknown as Record<string, unknown>)?.has_more === true;
      return { stockTakes, hasMore, page, pageSize };
    },
  });
}

export function useCreateStockTake() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<StockTake>) => unwrap(erpApi.createStockTake(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-takes'] }),
    onError: () => toast.error('Failed to create stock take'),
  });
}
