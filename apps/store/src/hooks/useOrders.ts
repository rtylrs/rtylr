import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi, StoreOrder } from '../services/storeService';

export interface OrderFilters {
  search?: string;
  status?: StoreOrder['status'];
  limit?: number;
  offset?: number;
}

export function useOrders(filters?: OrderFilters) {
  return useQuery({
    queryKey: ['store-orders', filters],
    queryFn: () => unwrap(storeApi.listOrders({
      search: filters?.search,
      status: filters?.status,
      limit: filters?.limit,
      offset: filters?.offset,
    })),
  });
}

export function useOrder(id: number | undefined) {
  return useQuery<StoreOrder | null>({
    queryKey: ['store-order', id],
    enabled: !!id,
    queryFn: () => unwrap(storeApi.getOrder(id!)),
  });
}

export function useUpdateOrderStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: number; status: StoreOrder['status'] }) =>
      unwrap(storeApi.updateOrderStatus(id, status)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['store-orders'] });
      qc.invalidateQueries({ queryKey: ['store-dashboard'] });
    },
  });
}
