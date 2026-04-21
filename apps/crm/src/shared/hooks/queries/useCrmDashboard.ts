import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { dashboardStatsSchema } from '../../schemas/dashboard';
import type { DashboardStats } from '../../schemas/dashboard';

export const useCrmDashboard = (): UseQueryResult<DashboardStats> =>
  useQuery<DashboardStats>({
    queryKey: ['crm-dashboard'],
    queryFn: async (): Promise<DashboardStats> => {
      const res = await unwrap(safe.get<unknown>('/crm/dashboard'));
      return dashboardStatsSchema.parse(res);
    },
  });

export type { DashboardStats } from '../../schemas/dashboard';
