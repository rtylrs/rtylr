import { createOneKitClient } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { RevenueServiceClient } from '@voxire/voxie/insights/revenue/v1/client';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type {
  RevenueSeriesResponse,
  RevenueByAppResponse,
  RevenueComparisonResponse,
} from './types';

const revenueClient = createOneKitClient(RevenueServiceClient, 'insights');

const majorToMinor = (value: number): number => Math.round(value * 100);

export const useRevenueSeries = (
  params?: DateRangeParams
): UseQueryResult<RevenueSeriesResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<RevenueSeriesResponse>({
    queryKey: ['insights-revenue', { from, to }],
    queryFn: async (): Promise<RevenueSeriesResponse> => {
      const res = await revenueClient.getRevenueOverview({ period: `${from},${to}` });
      const trend = res.overview?.trend ?? [];
      return {
        series: trend.map((point) => ({
          date: point.date ?? '',
          amountMinorUnits: majorToMinor(point.amount ?? 0),
        })),
      };
    },
  });
};

export const useRevenueByApp = (params?: DateRangeParams): UseQueryResult<RevenueByAppResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<RevenueByAppResponse>({
    queryKey: ['insights-revenue-by-app', { from, to }],
    queryFn: async (): Promise<RevenueByAppResponse> => {
      const res = await revenueClient.getRevenueByApp({ period: `${from},${to}` });
      return {
        byApp: (res.apps ?? []).map((entry) => ({
          app: entry.app ?? '',
          amountMinorUnits: majorToMinor(entry.amount ?? 0),
          pct: entry.percentage ?? 0,
        })),
      };
    },
  });
};

export const useRevenueComparison = (
  params?: DateRangeParams
): UseQueryResult<RevenueComparisonResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<RevenueComparisonResponse>({
    queryKey: ['insights-revenue-comparison', { from, to }],
    queryFn: async (): Promise<RevenueComparisonResponse> => {
      const res = await revenueClient.getRevenueComparison({
        period: `${from},${to}`,
        compareTo: 'previous_period',
      });
      const first = (res.comparisons ?? [])[0];
      return {
        thisPeriod: majorToMinor(first?.current ?? 0),
        lastPeriod: majorToMinor(first?.previous ?? 0),
        deltaPct: first?.change ?? 0,
      };
    },
  });
};
