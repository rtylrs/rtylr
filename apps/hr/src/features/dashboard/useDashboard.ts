import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { hrApi } from '../../shared/services/hrApi';
import { hrDashboardStatsSchema } from '../../shared/schemas/dashboard';

export function useHRDashboard() {
  return useQuery({
    queryKey: ['hr-dashboard'],
    queryFn: async () => {
      const res = await unwrap(hrApi.getDashboard());
      return hrDashboardStatsSchema.parse(res);
    },
  });
}
