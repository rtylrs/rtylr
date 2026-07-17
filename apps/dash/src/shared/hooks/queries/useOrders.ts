import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { dashOrderSchema } from '../../schemas/orders';

export type { HubOrder } from '../../schemas/orders';

export const useOrders = () => {
  return useQuery({
    queryKey: ['dash-orders'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/pos/orders/v1/orders'));
      const raw = res?.orders ?? res;
      return z.array(dashOrderSchema).parse(Array.isArray(raw) ? raw : []);
    },
    refetchInterval: 15_000,
  });
};
