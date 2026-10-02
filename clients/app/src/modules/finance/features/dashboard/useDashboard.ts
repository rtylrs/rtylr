import { oneKitResult, unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { financeDashboardClient } from '../../shared/lib/clients';
import type { DashboardStats } from '../../shared/services/financeApi';
import { dashboardStatsSchema } from '../../shared/schemas/dashboard';

export function useFinanceDashboard() {
  return useQuery({
    queryKey: ['finance', 'finance-dashboard'],
    queryFn: async (): Promise<DashboardStats> => {
      const res = await unwrap(oneKitResult(() => financeDashboardClient.getFinanceDashboard({})));
      return dashboardStatsSchema.parse(res) as DashboardStats;
    },
  });
}
