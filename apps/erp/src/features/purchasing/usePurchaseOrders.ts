import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { CreatePurchaseOrderRequest } from '@voxire/contracts/erp/v1/purchase_orders';
import { erpApi } from '../../shared/services/erpApi';

export function usePurchaseOrders() {
  return useQuery({
    queryKey: ['purchase-orders'],
    queryFn: () => unwrap(erpApi.listPurchaseOrders()),
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
