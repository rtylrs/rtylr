import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { DashboardServiceClient } from '@voxire/voxie/insights/dashboard/v1/client';
import type { InsightsDashboardResponse } from './types';

const dashboardClient = createOneKitClient(DashboardServiceClient, 'insights');

export const useInsightsDashboard = (): UseQueryResult<InsightsDashboardResponse> => {
  return useQuery<InsightsDashboardResponse>({
    queryKey: ['insights-dashboard'],
    refetchInterval: 60000,
    queryFn: async (): Promise<InsightsDashboardResponse> => {
      const res = await dashboardClient.getDashboardSummary({});
      return {
        revenue: {
          value: toFiniteNumber(res.revenue?.value),
          deltaPct: toFiniteNumber(res.revenue?.deltaPct),
        },
        customers: {
          value: toFiniteNumber(res.customers?.value),
          deltaPct: toFiniteNumber(res.customers?.deltaPct),
        },
        orders: {
          value: toFiniteNumber(res.orders?.value),
          deltaPct: toFiniteNumber(res.orders?.deltaPct),
        },
        margin: {
          value: toFiniteNumber(res.margin?.value),
          deltaPct: toFiniteNumber(res.margin?.deltaPct),
        },
      };
    },
  });
};
