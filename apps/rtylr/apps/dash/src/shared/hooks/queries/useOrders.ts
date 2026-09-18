import { z } from 'zod';
import i18next from 'i18next';
import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { rest } from '../../lib/clients';
import { dashOrderSchema } from '../../schemas/orders';

export type { HubOrder } from '../../schemas/orders';

interface UseOrdersParams {
  limit?: number;
  offset?: number;
  status?: string;
}

export const useOrders = (params: UseOrdersParams = {}) => {
  const { limit = 50, offset = 0, status } = params;
  return useQuery({
    queryKey: ['dash-orders', { limit, offset, status }],
    queryFn: async () => {
      const searchParams = new URLSearchParams({
        limit: String(limit),
        offset: String(offset),
      });
      if (status && status !== 'all') {
        searchParams.set('status', status);
      }
      const res = await unwrap(rest.get(`/pos/orders/v1/orders?${searchParams.toString()}`));
      const raw = res?.orders ?? res;
      return z.array(dashOrderSchema).parse(Array.isArray(raw) ? raw : []);
    },
    refetchOnWindowFocus: true,
  });
};

export const useUpdateOrderStatus = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, status }: { id: string; status: string }) => {
      return unwrap(rest.put(`/pos/orders/v1/orders/${id}/status`, { status }));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dash-orders'] });
      toast.success(i18next.t('common.toasts.orderStatusUpdated'));
    },
    onError: () => {
      toast.error(i18next.t('common.toasts.failedUpdateOrderStatus'));
    },
  });
};
