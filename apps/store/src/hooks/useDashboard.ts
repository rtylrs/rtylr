import { useQuery } from '@tanstack/react-query';
import { storeApi } from '../services/storeService';

export function useDashboardStats() {
  return useQuery({
    queryKey: ['store-dashboard'],
    queryFn: async () => {
      const r = await storeApi.getDashboardStats();
      return r.data ?? null;
    },
  });
}
