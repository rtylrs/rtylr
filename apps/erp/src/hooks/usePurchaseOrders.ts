import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi } from '../services/erpApi';

export function usePurchaseOrders() {
  return useQuery({
    queryKey: ['purchase-orders'],
    queryFn: () => erpApi.listPurchaseOrders().then((r) => r.data || []),
  });
}

export function useCreatePurchaseOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => erpApi.createPurchaseOrder(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
  });
}

export function useUpdatePurchaseOrderStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: string }) => erpApi.updatePurchaseOrderStatus(id, status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['purchase-orders'] }),
  });
}
