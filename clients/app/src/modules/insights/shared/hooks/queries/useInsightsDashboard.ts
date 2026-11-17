import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { DashboardServiceClient } from '@voxire/voxie/insights/dashboard/v1/client';
import type { InsightsDashboardResponse, SalesSummaryResponse, WorkSummaryResponse } from './types';

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

// Business-wide Work (Flow) totals and per-project progress, so a service
// business sees its delivery on the overview, not only POS and invoice cards.
export const useWorkSummary = (): UseQueryResult<WorkSummaryResponse> => {
  return useQuery<WorkSummaryResponse>({
    queryKey: ['insights-dashboard-work'],
    refetchInterval: 60000,
    queryFn: async (): Promise<WorkSummaryResponse> => {
      const [stats, roadmap] = await Promise.all([
        dashboardClient.getInsights({}),
        dashboardClient.getRoadmap({}),
      ]);
      return {
        totalTasks: toFiniteNumber(stats.totalTasks),
        completedTasks: toFiniteNumber(stats.completedTasks),
        overdueTasks: toFiniteNumber(stats.overdueTasks),
        projects: (roadmap.projects ?? [])
          .map((p) => ({
            id: p.projectId ?? '',
            name: p.projectName ?? '',
            totalTasks: toFiniteNumber(p.totalTasks),
            completedTasks: toFiniteNumber(p.completedTasks),
            progressPct: toFiniteNumber(p.progressPercent),
          }))
          .filter((p) => p.totalTasks > 0),
      };
    },
  });
};

// CRM open pipeline and the deal value won this month, compared with the same
// span last month.
export const useSalesSummary = (): UseQueryResult<SalesSummaryResponse> => {
  return useQuery<SalesSummaryResponse>({
    queryKey: ['insights-dashboard-sales'],
    refetchInterval: 60000,
    queryFn: async (): Promise<SalesSummaryResponse> => {
      const res = await dashboardClient.getSalesSummary({});
      return {
        openPipelineValue: toFiniteNumber(res.openPipelineValue),
        openDealCount: toFiniteNumber(res.openDealCount),
        won: {
          value: toFiniteNumber(res.won?.value),
          deltaPct: toFiniteNumber(res.won?.deltaPct),
        },
        recentWonDealCount: toFiniteNumber(res.recentWonDealCount),
      };
    },
  });
};
