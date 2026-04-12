import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeApi, DashboardStats } from '../../shared/services/financeApi';
import { dashboardStatsSchema } from '../../shared/schemas/dashboard';

export function useFinanceDashboard() {
  return useQuery({
    queryKey: ['finance-dashboard'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getDashboard());
      return dashboardStatsSchema.parse(res) as DashboardStats;
    },
  });
}
