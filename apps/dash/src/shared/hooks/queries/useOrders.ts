import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { orderClient, ordersCompatibilityClient } from '../../lib/clients';
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
      const res = await ordersCompatibilityClient.listOrders({
        limit,
        offset,
        status: status && status !== 'all' ? status : undefined,
      });
      return z.array(dashOrderSchema).parse(res.orders ?? []);
    },
    refetchOnWindowFocus: true,
  });
};

export const useUpdateOrderStatus = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, status }: { id: string; status: string }) => {
      return orderClient.updateOrderStatus({ orderId: id, body: { status } });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dash-orders'] });
      toast.success('Order status updated');
    },
    onError: () => {
      toast.error('Failed to update order status');
    },
  });
};
