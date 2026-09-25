import { z } from 'zod';
import i18next from 'i18next';
import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { rest } from '../../lib/clients';
import { dashOrderSchema, type HubOrder } from '../../schemas/orders';

export type { HubOrder } from '../../schemas/orders';

interface UseOrdersParams {
  limit?: number;
  offset?: number;
  status?: string;
}

export interface OrdersPage {
  orders: HubOrder[];
  total: number;
  hasMore: boolean;
  page: number;
  pageSize: number;
}

function positiveInteger(value: unknown): number | undefined {
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : undefined;
}

function nonNegativeInteger(value: unknown): number | undefined {
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed >= 0 ? parsed : undefined;
}

function responseRecord(value: unknown): Record<string, unknown> {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};
}

export const useOrders = (params: UseOrdersParams = {}) => {
  const { limit = 50, offset = 0, status } = params;
  return useQuery<OrdersPage>({
    queryKey: ['dash-orders', { limit, offset, status }],
    queryFn: async () => {
      const requestedPageSize = positiveInteger(limit);
      const pageSize = requestedPageSize ? Math.min(200, requestedPageSize) : 50;
      const normalizedOffset = nonNegativeInteger(offset) ?? 0;
      const page = Math.floor(normalizedOffset / pageSize) + 1;
      const searchParams = new URLSearchParams({
        page: String(page),
        page_size: String(pageSize),
      });
      if (status && status !== 'all') {
        searchParams.set('status', status);
      }
      const response = await unwrap(rest.get(`/pos/orders?${searchParams.toString()}`));
      const source = responseRecord(response);
      const rawOrders = Array.isArray(response)
        ? response
        : Array.isArray(source.orders)
          ? source.orders
          : [];
      const orders = z.array(dashOrderSchema).parse(rawOrders);
      const responsePage = positiveInteger(source.page) ?? page;
      const responsePageSize = positiveInteger(source.page_size ?? source.pageSize) ?? pageSize;
      const total = nonNegativeInteger(source.total) ?? orders.length;
      const hasMore =
        typeof source.has_more === 'boolean'
          ? source.has_more
          : typeof source.hasMore === 'boolean'
            ? source.hasMore
            : responsePage * responsePageSize < total;

      return { orders, total, hasMore, page: responsePage, pageSize: responsePageSize };
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
