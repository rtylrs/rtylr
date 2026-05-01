import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { CreatePurchaseOrderRequest } from '@voxire/contracts/erp/v1/purchase_orders';
import { erpApi, PurchaseOrder } from '../../shared/services/erpApi';
import { purchaseOrderSchema } from '../../shared/schemas/purchaseOrders';
import { unwrapList } from '../../shared/lib/unwrapList';

export function usePurchaseOrders() {
  return useQuery({
    queryKey: ['purchase-orders'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listPurchaseOrders());
      return z
        .array(purchaseOrderSchema)
        .parse(unwrapList(res, 'purchase_orders', 'orders')) as PurchaseOrder[];
    },
  });
}

export function useCreatePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePurchaseOrderRequest) => unwrap(erpApi.createPurchaseOrder(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
  });
}

export function useUpdatePurchaseOrderStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) =>
      unwrap(erpApi.updatePurchaseOrderStatus(id, status)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
  });
}

export function useApprovePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.approvePurchaseOrder(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
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
  });
}
