import { createOneKitClient } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
} from '@tanstack/react-query';
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

/**
 * A mutation request carrying the `Idempotency-Key` for the operator action
 * behind it.
 *
 * Both stock-transfer routes require the header (the server answers 428 without
 * it) and it is also what dedups a retry, so it identifies the action rather
 * than the request.
 */
type Keyed<TRequest> = TRequest & { idempotencyKey: string };

/**
 * Wraps a mutation so every `mutate`/`mutateAsync` call carries its own
 * `Idempotency-Key`, minted once when the operator acts and reused by every
 * attempt of that action.
 *
 * The key travels in the variables because that is the one thing react-query
 * hands back unchanged to `mutationFn` on each retry. A later, separate submit
 * mints a different key, so the server can still tell the two apart.
 */
function withIdempotencyKey<TData, TRequest>(
  mutation: UseMutationResult<TData, Error, Keyed<TRequest>>
) {
  return {
    ...mutation,
    mutate: (request: TRequest, options?: Parameters<typeof mutation.mutate>[1]) =>
      mutation.mutate({ ...request, idempotencyKey: crypto.randomUUID() }, options),
    mutateAsync: (request: TRequest, options?: Parameters<typeof mutation.mutateAsync>[1]) =>
      mutation.mutateAsync({ ...request, idempotencyKey: crypto.randomUUID() }, options),
  };
}

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
  const mutation = useMutation<
    CreateStockTransferResponse,
    Error,
    Keyed<CreateStockTransferRequest>
  >({
    // The generated client already accepts per-call RequestOptions, so the key
    // reaches the wire without a contract change. Without it the server answers
    // this POST with 428 and no draft is created.
    mutationFn: ({ idempotencyKey, ...data }) =>
      stockTransferClient.createStockTransfer(data, {
        headers: { 'Idempotency-Key': idempotencyKey },
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-transfers'] }),
    onError: () => {
      void qc.invalidateQueries({ queryKey: ['stock-transfers'] });
      toast.error('Failed to create stock transfer');
    },
  });

  return withIdempotencyKey(mutation);
}

export function useUpdateStockTransferStatus() {
  const qc = useQueryClient();
  const mutation = useMutation<
    UpdateStockTransferStatusResponse,
    Error,
    Keyed<UpdateStockTransferStatusRequest>
  >({
    mutationFn: ({ idempotencyKey, ...data }) =>
      stockTransferClient.updateStockTransferStatus(data, {
        headers: { 'Idempotency-Key': idempotencyKey },
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['stock-transfers'] });
      invalidateStockDerivedQueries(qc);
    },
    onError: () => {
      void qc.invalidateQueries({ queryKey: ['stock-transfers'] });
      invalidateStockDerivedQueries(qc);
      toast.error('Failed to update stock transfer');
    },
  });

  return withIdempotencyKey(mutation);
}
