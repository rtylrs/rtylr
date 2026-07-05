import { createOneKitClient } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { DashboardServiceClient } from '@voxire/voxie/insights/dashboard/v1/service_client';
import type { InsightsDashboardResponse } from './types';

const dashboardClient = createOneKitClient(DashboardServiceClient, 'insights');

export const useInsightsDashboard = (): UseQueryResult<InsightsDashboardResponse> => {
  return useQuery<InsightsDashboardResponse>({
    queryKey: ['insights-dashboard'],
    queryFn: async (): Promise<InsightsDashboardResponse> => {
      const res = await dashboardClient.getDashboardSummary({});
      return {
        revenue: {
          value: Number(res.revenue?.value ?? 0),
          deltaPct: res.revenue?.deltaPct ?? 0,
        },
        customers: {
          value: Number(res.customers?.value ?? 0),
          deltaPct: res.customers?.deltaPct ?? 0,
        },
        orders: {
          value: Number(res.orders?.value ?? 0),
          deltaPct: res.orders?.deltaPct ?? 0,
        },
        margin: {
          value: res.margin?.value ?? 0,
          deltaPct: res.margin?.deltaPct ?? 0,
        },
      };
    },
  });
};
