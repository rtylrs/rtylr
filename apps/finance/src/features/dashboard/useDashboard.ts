import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import type { DashboardStats } from '../../shared/services/financeApi';
import { dashboardStatsSchema } from '../../shared/schemas/dashboard';

export function useFinanceDashboard() {
  return useQuery({
    queryKey: ['finance-dashboard'],
    queryFn: async (): Promise<DashboardStats> => {
      const res = await unwrap(safe.get<unknown>('/finance/dashboard'));
      return dashboardStatsSchema.parse(res) as DashboardStats;
    },
  });
}
