import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type {
  RevenueSeriesResponse,
  RevenueByAppResponse,
  RevenueComparisonResponse,
} from './types';

export const useRevenueSeries = (
  params?: DateRangeParams
): UseQueryResult<RevenueSeriesResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<RevenueSeriesResponse>({
    queryKey: ['insights-revenue', { from, to }],
    queryFn: async (): Promise<RevenueSeriesResponse> => {
      const res = await unwrap(
        safe.get<RevenueSeriesResponse>('/insights/revenue', { params: { from, to } })
      );
      return res as RevenueSeriesResponse;
    },
  });
};

export const useRevenueByApp = (): UseQueryResult<RevenueByAppResponse> => {
  return useQuery<RevenueByAppResponse>({
    queryKey: ['insights-revenue-by-app'],
    queryFn: async (): Promise<RevenueByAppResponse> => {
      const res = await unwrap(safe.get<RevenueByAppResponse>('/insights/revenue/by-app'));
      return res as RevenueByAppResponse;
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
        safe.get<RevenueComparisonResponse>('/insights/revenue/comparison', {
          params: { from, to },
        })
      );
      return res as RevenueComparisonResponse;
    },
  });
};
