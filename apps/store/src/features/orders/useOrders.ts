import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { storeApi, StoreOrder } from '../../shared/services/storeService';
import { storeOrderSchema } from '../../shared/schemas/orders';

export interface OrderFilters {
  search?: string;
  status?: StoreOrder['status'];
  limit?: number;
  offset?: number;
}

export function useOrders(filters?: OrderFilters) {
  return useQuery({
    queryKey: ['store-orders', filters],
    queryFn: async () => {
      const res = await unwrap(
        storeApi.listOrders({
          search: filters?.search,
          status: filters?.status,
          limit: filters?.limit,
          offset: filters?.offset,
        })
      );
      const raw = (res as any)?.orders ?? res;
      return z.array(storeOrderSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useOrder(id: number | undefined) {
  return useQuery({
    queryKey: ['store-order', id],
    enabled: !!id,
    queryFn: async () => {
      const res = await unwrap(storeApi.getOrder(id!));
      return storeOrderSchema.parse(res);
    },
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
