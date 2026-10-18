import { createOneKitClient } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { PurchaseOrderServiceClient } from '@voxire/voxie/erp/purchase_order/v1/client';
import type {
  ApprovePurchaseOrderResponse,
  CreatePurchaseOrderRequest,
  CreatePurchaseOrderResponse,
  ListPurchaseOrdersResponse,
  PurchaseOrder,
  UpdatePurchaseOrderStatusRequest,
  UpdatePurchaseOrderStatusResponse,
} from '@voxire/voxie/erp/purchase_order/v1/types';
import { invalidateStockDerivedQueries } from '../../shared/lib/stockInvalidation';

const purchaseOrderClient = createOneKitClient(PurchaseOrderServiceClient, 'erp');

export type { CreatePurchaseOrderRequest, PurchaseOrder };

export interface UsePurchaseOrdersParams {
  page?: number;
  pageSize?: number;
  status?: string;
}

const EXPORT_PAGE_SIZE = 100;

export async function fetchAllPurchaseOrders(
  params: Pick<UsePurchaseOrdersParams, 'status'> = {}
): Promise<PurchaseOrder[]> {
  const purchaseOrders: PurchaseOrder[] = [];
  let page = 1;

  while (true) {
    const response = await purchaseOrderClient.listPurchaseOrders({
      status: params.status ?? '',
      page,
      pageSize: EXPORT_PAGE_SIZE,
    });
    const pageOrders = response.purchaseOrders ?? [];
    purchaseOrders.push(...pageOrders);

    const hasMore = response.hasMore ?? pageOrders.length === EXPORT_PAGE_SIZE;
    if (!hasMore || pageOrders.length === 0) return purchaseOrders;
    page += 1;
  }
}

export function usePurchaseOrders(params: UsePurchaseOrdersParams = {}) {
  const { page = 1, pageSize = 50, status = '' } = params;
  return useQuery<ListPurchaseOrdersResponse>({
    queryKey: ['purchase-orders', { page, pageSize, status }],
    queryFn: () => purchaseOrderClient.listPurchaseOrders({ page, pageSize, status }),
  });
}

/**
 * Detail fetch for one purchase order. The list response carries no line
 * items, so screens that need the lines (Receiving) must load the order
 * through getPurchaseOrder.
 */
export function usePurchaseOrder(id: string | undefined) {
  return useQuery<PurchaseOrder | null>({
    queryKey: ['purchase-orders', 'detail', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      const response = await purchaseOrderClient.getPurchaseOrder({ id });
      return response.purchaseOrder ?? null;
    },
  });
}

export function useAllPurchaseOrders() {
  return useQuery<PurchaseOrder[]>({
    queryKey: ['purchase-orders', 'all'],
    queryFn: () => fetchAllPurchaseOrders(),
  });
}

// Mutations leave error reporting to the screen that triggered them, which
// knows the context; a hook-level toast on top of that showed every failure
// twice.
export function useCreatePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation<CreatePurchaseOrderResponse, Error, CreatePurchaseOrderRequest>({
    mutationFn: (data) => purchaseOrderClient.createPurchaseOrder(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
  });
}

export function useUpdatePurchaseOrderStatus() {
  const qc = useQueryClient();
  return useMutation<UpdatePurchaseOrderStatusResponse, Error, UpdatePurchaseOrderStatusRequest>({
    mutationFn: (data) => purchaseOrderClient.updatePurchaseOrderStatus(data),
    onSuccess: (_response, request) => {
      void qc.invalidateQueries({ queryKey: ['purchase-orders'] });
      // Receiving adds stock, so every stock-derived view is stale.
      if (request.body?.status === 'received') invalidateStockDerivedQueries(qc);
    },
  });
}

export function useApprovePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation<ApprovePurchaseOrderResponse, Error, string>({
    mutationFn: (id) => purchaseOrderClient.approvePurchaseOrder({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
  });
}
