import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type {
  RevenueSeriesResponse,
  RevenueByAppResponse,
  RevenueComparisonResponse,
} from './types';
import {
  mapRevenueByApp,
  mapRevenueComparison,
  mapRevenueOverview,
} from '../../utils/onekitMappers';

export const useRevenueSeries = (
  params?: DateRangeParams
): UseQueryResult<RevenueSeriesResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<RevenueSeriesResponse>({
    queryKey: ['insights-revenue', { from, to }],
    queryFn: async (): Promise<RevenueSeriesResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/revenue/v1/overview', {
          params: { period: `${from},${to}` },
        })
      );
      return mapRevenueOverview(res as Record<string, unknown>);
    },
  });
};

export const useRevenueByApp = (): UseQueryResult<RevenueByAppResponse> => {
  return useQuery<RevenueByAppResponse>({
    queryKey: ['insights-revenue-by-app'],
    queryFn: async (): Promise<RevenueByAppResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/revenue/v1/by-app')
      );
      return mapRevenueByApp(res as Record<string, unknown>);
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
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/revenue/v1/comparison', {
          params: { period: `${from},${to}`, compare_to: 'previous_period' },
        })
      );
      return mapRevenueComparison(res as Record<string, unknown>);
    },
  });
};
