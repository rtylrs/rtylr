import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { StockTakeServiceClient } from '@voxire/voxie/erp/stock_take/v1/client';
import type {
  CreateStockTakeRequest,
  CreateStockTakeResponse,
  ListStockTakesResponse,
  StockTake,
  UpdateStockTakeRequest,
  UpdateStockTakeResponse,
} from '@voxire/voxie/erp/stock_take/v1/types';
import { invalidateStockDerivedQueries } from '../../shared/lib/stockInvalidation';

const stockTakeClient = createOneKitClient(StockTakeServiceClient, 'erp');

export type {
  CreateStockTakeRequest as CreateStockTakeInput,
  StockTake,
  UpdateStockTakeRequest as UpdateStockTakeInput,
};

export interface UseStockTakesParams {
  page?: number;
  pageSize?: number;
}

export function useStockTakes(params: UseStockTakesParams = {}) {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;
  return useQuery<ListStockTakesResponse>({
    queryKey: ['stock-takes', { page, pageSize }],
    queryFn: () => stockTakeClient.listStockTakes({ page, pageSize }),
  });
}

export function useCreateStockTake() {
  const qc = useQueryClient();
  return useMutation<CreateStockTakeResponse, Error, CreateStockTakeRequest>({
    mutationFn: (data) => stockTakeClient.createStockTake(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-takes'] }),
    onError: () => {
      void qc.invalidateQueries({ queryKey: ['stock-takes'] });
      toast.error('Failed to create stock take');
    },
  });
}

export function useUpdateStockTake() {
  const qc = useQueryClient();
  return useMutation<UpdateStockTakeResponse, Error, UpdateStockTakeRequest>({
    mutationFn: (data) => stockTakeClient.updateStockTake(data),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['stock-takes'] });
      invalidateStockDerivedQueries(qc);
    },
    onError: () => {
      void qc.invalidateQueries({ queryKey: ['stock-takes'] });
      invalidateStockDerivedQueries(qc);
      toast.error('Failed to update stock take');
    },
  });
}
