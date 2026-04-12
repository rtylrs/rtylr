import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi } from '../../shared/services/storeService';
import { storeDashboardStatsSchema } from '../../shared/schemas/dashboard';

export function useDashboardStats() {
  return useQuery({
    queryKey: ['store-dashboard'],
    queryFn: async () => {
      const res = await unwrap(storeApi.getDashboardStats());
      return storeDashboardStatsSchema.parse(res);
    },
  });
}
