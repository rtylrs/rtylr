import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { CreatePurchaseOrderRequest } from '@voxire/contracts/erp/v1/purchase_orders';
import { erpApi, PurchaseOrder } from '../../shared/services/erpApi';
import { purchaseOrderSchema } from '../../shared/schemas/purchaseOrders';

export function usePurchaseOrders() {
  return useQuery({
    queryKey: ['purchase-orders'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listPurchaseOrders());
      return z.array(purchaseOrderSchema).parse(Array.isArray(res) ? res : []) as PurchaseOrder[];
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
