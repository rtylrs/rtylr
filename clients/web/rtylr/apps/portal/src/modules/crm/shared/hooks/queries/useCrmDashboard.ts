import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { crmApi } from '../../services/crmApi';
import { dashboardStatsSchema } from '../../schemas/dashboard';
import type { DashboardStats } from '../../schemas/dashboard';

export const useCrmDashboard = (): UseQueryResult<DashboardStats> =>
  useQuery<DashboardStats>({
    queryKey: ['crm-dashboard'],
    queryFn: async (): Promise<DashboardStats> => {
      const res = await unwrap(crmApi.getDashboard());
      return dashboardStatsSchema.parse(res);
    },
  });

export type { DashboardStats } from '../../schemas/dashboard';
