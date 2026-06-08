import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

// No proto type yet - define locally.
export interface StockTake {
  id: number;
  name: string;
  warehouseId: number;
  warehouseName: string;
  status: 'Draft' | 'In Progress' | 'Completed';
  itemsCounted: number;
  totalItems: number;
  variance: number;
  createdAt: string;
}

export interface CreateStockTakeInput {
  name: string;
  warehouseId: number;
  notes?: string;
  items: Array<{ productId: number; expectedQty: number; countedQty: number }>;
}

interface ListStockTakesResponse {
  stockTakes: StockTake[];
}

export const useListStockTakes = (): UseQueryResult<StockTake[]> => {
  return useQuery<StockTake[]>({
    queryKey: ['erp-stock-takes'],
    queryFn: async (): Promise<StockTake[]> => {
      const res = await unwrap(safe.get<ListStockTakesResponse>('/erp/stock-takes'));
      return (res as ListStockTakesResponse).stockTakes ?? [];
    },
  });
};

export const useCreateStockTake = (): UseMutationResult<StockTake, Error, CreateStockTakeInput> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateStockTakeInput): Promise<StockTake> => {
      const res = await unwrap(
        safe.post<StockTake>('/erp/stock-takes', {
          warehouseId: data.warehouseId,
          notes: data.notes,
          items: data.items,
        })
      );
      return res as StockTake;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-stock-takes'] });
      toast.success(t('erp.stockTakeCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
