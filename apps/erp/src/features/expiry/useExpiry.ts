import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, ExpiryItem } from '../../shared/services/erpApi';
import { expiryItemSchema } from '../../shared/schemas/expiry';

export function useExpiry() {
  return useQuery<ExpiryItem[]>({
    queryKey: ['expiry'],
    queryFn: async () => {
      const res = await unwrap(erpApi.getExpiryItems());
      return z.array(expiryItemSchema).parse(Array.isArray(res) ? res : []) as ExpiryItem[];
    },
  });
}
