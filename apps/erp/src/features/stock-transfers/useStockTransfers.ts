import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { StockTransferServiceClient } from '@voxire/voxie/erp/stock_transfer/v1/client';
import type {
  CreateStockTransferRequest,
  CreateStockTransferResponse,
  ListStockTransfersResponse,
  StockTransfer,
  UpdateStockTransferStatusRequest,
  UpdateStockTransferStatusResponse,
} from '@voxire/voxie/erp/stock_transfer/v1/types';
import { invalidateStockDerivedQueries } from '../../shared/lib/stockInvalidation';

const stockTransferClient = createOneKitClient(StockTransferServiceClient, 'erp');

export type {
  CreateStockTransferRequest as CreateStockTransferInput,
  StockTransfer,
  UpdateStockTransferStatusRequest as UpdateStockTransferStatusInput,
};

export interface UseStockTransfersParams {
  page?: number;
  pageSize?: number;
}

export function useStockTransfers(params: UseStockTransfersParams = {}) {
  const page = params.page ?? 1;
  const pageSize = params.pageSize ?? 50;
  return useQuery<ListStockTransfersResponse>({
    queryKey: ['stock-transfers', { page, pageSize }],
    queryFn: () => stockTransferClient.listStockTransfers({ page, pageSize }),
  });
}

export function useCreateStockTransfer() {
  const qc = useQueryClient();
  return useMutation<CreateStockTransferResponse, Error, CreateStockTransferRequest>({
    mutationFn: (data) => stockTransferClient.createStockTransfer(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-transfers'] }),
    onError: () => toast.error('Failed to create stock transfer'),
  });
}

export function useUpdateStockTransferStatus() {
  const qc = useQueryClient();
  return useMutation<UpdateStockTransferStatusResponse, Error, UpdateStockTransferStatusRequest>({
    mutationFn: (data) => stockTransferClient.updateStockTransferStatus(data),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['stock-transfers'] });
      invalidateStockDerivedQueries(qc);
    },
    onError: () => toast.error('Failed to update stock transfer'),
  });
}
