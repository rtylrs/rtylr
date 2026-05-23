import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import type { CreatePurchaseOrderRequest } from '@voxire/contracts/erp/v1/purchase_orders';
import { erpApi, PurchaseOrder } from '../../shared/services/erpApi';
import { purchaseOrderSchema } from '../../shared/schemas/purchaseOrders';
import { unwrapList } from '../../shared/lib/unwrapList';

interface UsePurchaseOrdersParams {
  page?: number;
  pageSize?: number;
}

export interface PurchaseOrdersResult {
  orders: PurchaseOrder[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

export function usePurchaseOrders(params: UsePurchaseOrdersParams = {}) {
  const { page = 1, pageSize = 50 } = params;
  return useQuery<PurchaseOrdersResult>({
    queryKey: ['purchase-orders', { page, pageSize }],
    queryFn: async (): Promise<PurchaseOrdersResult> => {
      const res = await unwrap(erpApi.listPurchaseOrders({ page, page_size: pageSize }));
      const raw = res as unknown as Record<string, unknown>;
      const orders = z
        .array(purchaseOrderSchema)
        .parse(unwrapList(res, 'purchase_orders', 'orders')) as PurchaseOrder[];
      return {
        orders,
        hasMore: Boolean(raw.has_more),
        page: Number(raw.page ?? page),
        pageSize: Number(raw.page_size ?? pageSize),
      };
    },
  });
}

export function useCreatePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePurchaseOrderRequest) => unwrap(erpApi.createPurchaseOrder(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
    onError: () => toast.error('Failed to create purchase order'),
  });
}

export function useUpdatePurchaseOrderStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) =>
      unwrap(erpApi.updatePurchaseOrderStatus(id, status)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
    onError: () => toast.error('Failed to update purchase order status'),
  });
}

export function useApprovePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.approvePurchaseOrder(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
    onError: () => toast.error('Failed to approve purchase order'),
  });
}

export function useSendPurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    // Send transitions a draft PO into 'sent' status. Backend hangs
    // off updatePurchaseOrderStatus for now — when a dedicated
    // /send endpoint is added (with the supplier email side-effect),
    // swap the mutationFn here and the call sites stay unchanged.
    mutationFn: (id: number) => unwrap(erpApi.updatePurchaseOrderStatus(id, 'sent')),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
    onError: () => toast.error('Failed to send purchase order'),
  });
}
