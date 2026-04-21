import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import type { InsightsDashboardResponse } from './types';

export const useInsightsDashboard = (): UseQueryResult<InsightsDashboardResponse> => {
  return useQuery<InsightsDashboardResponse>({
    queryKey: ['insights-dashboard'],
    queryFn: async (): Promise<InsightsDashboardResponse> => {
      const res = await unwrap(safe.get<InsightsDashboardResponse>('/insights/dashboard'));
      return res as InsightsDashboardResponse;
    },
  });
};
