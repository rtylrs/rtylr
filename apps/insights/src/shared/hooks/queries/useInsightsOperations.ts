import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type {
  OperationsKPIResponse,
  HeadcountResponse,
  ShiftMetricsResponse,
  TimeOffResponse,
  UtilizationResponse,
} from './types';
import {
  mapHeadcountMetrics,
  mapOperationsSummary,
  mapShiftMetrics,
  mapTimeOffMetrics,
  mapUtilizationMetrics,
} from '../../utils/onekitMappers';

export const useOperationsKPIs = (): UseQueryResult<OperationsKPIResponse> => {
  return useQuery<OperationsKPIResponse>({
    queryKey: ['insights-operations'],
    queryFn: async (): Promise<OperationsKPIResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/operation/v1/summary')
      );
      return mapOperationsSummary(res as Record<string, unknown>);
    },
  });
};

export const useHeadcount = (): UseQueryResult<HeadcountResponse> => {
  return useQuery<HeadcountResponse>({
    queryKey: ['insights-operations-headcount'],
    queryFn: async (): Promise<HeadcountResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/operation/v1/headcount')
      );
      return mapHeadcountMetrics(res as Record<string, unknown>);
    },
  });
};

export const useShiftMetrics = (params?: DateRangeParams): UseQueryResult<ShiftMetricsResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<ShiftMetricsResponse>({
    queryKey: ['insights-operations-shifts', { from, to }],
    queryFn: async (): Promise<ShiftMetricsResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/operation/v1/shifts', {
          params: { period: `${from},${to}` },
        })
      );
      return mapShiftMetrics(res as Record<string, unknown>);
    },
  });
};

export const useTimeOff = (): UseQueryResult<TimeOffResponse> => {
  return useQuery<TimeOffResponse>({
    queryKey: ['insights-operations-time-off'],
    queryFn: async (): Promise<TimeOffResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/operation/v1/time-off')
      );
      return mapTimeOffMetrics(res as Record<string, unknown>);
    },
  });
};

export const useUtilization = (): UseQueryResult<UtilizationResponse> => {
  return useQuery<UtilizationResponse>({
    queryKey: ['insights-operations-utilization'],
    queryFn: async (): Promise<UtilizationResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/operation/v1/utilization')
      );
      return mapUtilizationMetrics(res as Record<string, unknown>);
    },
  });
};
