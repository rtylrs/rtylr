import { useQuery } from '@tanstack/react-query';
import { erpApi, ExpiryItem } from '../services/erpApi';

export function useExpiry() {
  return useQuery<ExpiryItem[]>({
    queryKey: ['expiry'],
    queryFn: async () => {
      const res = await erpApi.getExpiryItems();
      return res.data ?? [];
    },
  });
}
