import { useQuery } from '@tanstack/react-query';
import { erpApi, ExpiryItem } from '../services/erpApi';
import { mockExpiryItems } from '../lib/mockData';

export function useExpiry() {
  return useQuery<ExpiryItem[]>({
    queryKey: ['expiry'],
    queryFn: async () => {
      try {
        const res = await erpApi.getExpiryItems();
        if (res.data && res.data.length) return res.data;
        return mockExpiryItems;
      } catch {
        return mockExpiryItems;
      }
    },
  });
}
