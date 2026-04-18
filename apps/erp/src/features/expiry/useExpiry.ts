import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ExpiryItem } from '../../shared/services/erpApi';
import { expiryItemSchema } from '../../shared/schemas/expiry';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useExpiry() {
  return useQuery<ExpiryItem[]>({
    queryKey: ['expiry'],
    queryFn: async () => {
      const res = await unwrap(erpApi.getExpiryItems());
      return z.array(expiryItemSchema).parse(unwrapList(res, 'items', 'expiry')) as ExpiryItem[];
    },
  });
}
