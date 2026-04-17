import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { hubOrderSchema } from '../../schemas/orders';

export type { HubOrder } from '../../schemas/orders';

export const useOrders = () => {
  return useQuery({
    queryKey: ['hub-orders'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/store/orders'));
      const raw = res?.orders ?? res;
      return z.array(hubOrderSchema).parse(Array.isArray(raw) ? raw : []);
    },
    staleTime: 15_000,
  });
};
