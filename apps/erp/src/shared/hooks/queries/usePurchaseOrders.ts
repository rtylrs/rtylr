import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { PurchaseOrder } from '@voxire/contracts/erp/purchase_order/v1/types';
import type { ListPurchaseOrdersResponse } from '@voxire/contracts/erp/purchase_order/v1/list_purchase_orders';
import type {
  CreatePurchaseOrderRequest,
  CreatePurchaseOrderResponse,
} from '@voxire/contracts/erp/purchase_order/v1/create_purchase_order';
import type {
  UpdatePurchaseOrderStatusRequest,
  UpdatePurchaseOrderStatusResponse,
} from '@voxire/contracts/erp/purchase_order/v1/update_purchase_order_status';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface ListPurchaseOrdersParams {
  status?: string;
  page?: number;
  pageSize?: number;
}

export interface PurchaseOrderListResult {
  purchaseOrders: PurchaseOrder[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export const useListPurchaseOrders = (
  params: ListPurchaseOrdersParams = {}
): UseQueryResult<PurchaseOrderListResult> => {
  const { status, page = 1, pageSize = 50 } = params;
  return useQuery<PurchaseOrderListResult>({
    queryKey: ['erp-purchase-orders', 'filter', { status, page, pageSize }],
    queryFn: async (): Promise<PurchaseOrderListResult> => {
      const res = await unwrap(
        safe.get<unknown>('/erp/purchase-orders', {
          params: { status, page, page_size: pageSize },
        })
      );
      const raw = res as Record<string, unknown>;
      const purchaseOrders = (raw.purchaseOrders ?? []) as PurchaseOrder[];
      return {
        purchaseOrders,
        hasMore: Boolean(raw.hasMore),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.pageSize ?? pageSize),
      };
    },
  });
};

export const useGetPurchaseOrder = (
  id: number | undefined
): UseQueryResult<PurchaseOrder | null> => {
  return useQuery<PurchaseOrder | null>({
    queryKey: ['erp-purchase-order', id],
    enabled: !!id,
    queryFn: async (): Promise<PurchaseOrder | null> => {
      if (!id) return null;
      const res = await unwrap(
        safe.get<{ purchaseOrder: PurchaseOrder } | PurchaseOrder>(`/erp/purchase-orders/${id}`)
      );
      const maybeEnvelope = res as { purchaseOrder?: PurchaseOrder };
      if (maybeEnvelope && typeof maybeEnvelope === 'object' && 'purchaseOrder' in maybeEnvelope) {
        return maybeEnvelope.purchaseOrder ?? null;
      }
      return res as PurchaseOrder;
    },
  });
};

export const useCreatePurchaseOrder = (): UseMutationResult<
  CreatePurchaseOrderResponse,
  Error,
  CreatePurchaseOrderRequest
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreatePurchaseOrderRequest): Promise<CreatePurchaseOrderResponse> => {
      const res = await unwrap(
        safe.post<CreatePurchaseOrderResponse>('/erp/purchase-orders', data)
      );
      return res as CreatePurchaseOrderResponse;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-purchase-orders'] });
      toast.success(t('erp.purchaseOrderCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdatePurchaseOrderStatus = (): UseMutationResult<
  UpdatePurchaseOrderStatusResponse,
  Error,
  UpdatePurchaseOrderStatusRequest
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: UpdatePurchaseOrderStatusRequest
    ): Promise<UpdatePurchaseOrderStatusResponse> => {
      const res = await unwrap(
        safe.put<UpdatePurchaseOrderStatusResponse>('/erp/purchase-orders/status', {
          id: data.id,
          status: data.status,
        })
      );
      return res as UpdatePurchaseOrderStatusResponse;
    },
    onSuccess: (
      _res: UpdatePurchaseOrderStatusResponse,
      variables: UpdatePurchaseOrderStatusRequest
    ): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-purchase-orders'] });
      queryClient.invalidateQueries({ queryKey: ['erp-purchase-order', variables.id] });
      toast.success(t('erp.purchaseOrderStatusUpdated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useApprovePurchaseOrder = (): UseMutationResult<
  { message: string },
  Error,
  number
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: number): Promise<{ message: string }> => {
      const res = await unwrap(
        safe.post<{ message: string }>(`/erp/purchase-orders/${id}/approve`, {})
      );
      return res as { message: string };
    },
    onSuccess: (_res: { message: string }, id: number): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-purchase-orders'] });
      queryClient.invalidateQueries({ queryKey: ['erp-purchase-order', id] });
      toast.success(t('erp.purchaseOrderApproved'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
