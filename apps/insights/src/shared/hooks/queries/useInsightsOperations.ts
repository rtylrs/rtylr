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

export const useOperationsKPIs = (): UseQueryResult<OperationsKPIResponse> => {
  return useQuery<OperationsKPIResponse>({
    queryKey: ['insights-operations'],
    queryFn: async (): Promise<OperationsKPIResponse> => {
      const res = await unwrap(safe.get<OperationsKPIResponse>('/insights/operations'));
      return res as OperationsKPIResponse;
    },
  });
};

export const useHeadcount = (): UseQueryResult<HeadcountResponse> => {
  return useQuery<HeadcountResponse>({
    queryKey: ['insights-operations-headcount'],
    queryFn: async (): Promise<HeadcountResponse> => {
      const res = await unwrap(safe.get<HeadcountResponse>('/insights/operations/headcount'));
      return res as HeadcountResponse;
    },
  });
};

export const useShiftMetrics = (params?: DateRangeParams): UseQueryResult<ShiftMetricsResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<ShiftMetricsResponse>({
    queryKey: ['insights-operations-shifts', { from, to }],
    queryFn: async (): Promise<ShiftMetricsResponse> => {
      const res = await unwrap(
        safe.get<ShiftMetricsResponse>('/insights/operations/shifts', { params: { from, to } })
      );
      return res as ShiftMetricsResponse;
    },
  });
};

export const useTimeOff = (): UseQueryResult<TimeOffResponse> => {
  return useQuery<TimeOffResponse>({
    queryKey: ['insights-operations-time-off'],
    queryFn: async (): Promise<TimeOffResponse> => {
      const res = await unwrap(safe.get<TimeOffResponse>('/insights/operations/time-off'));
      return res as TimeOffResponse;
    },
  });
};

export const useUtilization = (): UseQueryResult<UtilizationResponse> => {
  return useQuery<UtilizationResponse>({
    queryKey: ['insights-operations-utilization'],
    queryFn: async (): Promise<UtilizationResponse> => {
      const res = await unwrap(safe.get<UtilizationResponse>('/insights/operations/utilization'));
      return res as UtilizationResponse;
    },
  });
};
