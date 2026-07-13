import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { OperationServiceClient } from '@voxire/voxie/insights/operation/v1/client';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type {
  OperationsKPIResponse,
  HeadcountResponse,
  ShiftMetricsResponse,
  TimeOffResponse,
  UtilizationResponse,
} from './types';

const operationClient = createOneKitClient(OperationServiceClient, 'insights');

export const useOperationsKPIs = (): UseQueryResult<OperationsKPIResponse> => {
  return useQuery<OperationsKPIResponse>({
    queryKey: ['insights-operations'],
    queryFn: async (): Promise<OperationsKPIResponse> => {
      const res = await operationClient.getOperationsSummary({});
      return {
        activeStaff: toFiniteNumber(res.activeStaff),
        shiftsThisWeek: toFiniteNumber(res.shiftsThisWeek),
        pendingTimeOff: toFiniteNumber(res.pendingTimeOff),
        openLeaveRequests: toFiniteNumber(res.openLeaveRequests),
      };
    },
  });
};

export const useHeadcount = (params?: DateRangeParams): UseQueryResult<HeadcountResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<HeadcountResponse>({
    queryKey: ['insights-operations-headcount', { from, to }],
    queryFn: async (): Promise<HeadcountResponse> => {
      const res = await operationClient.getHeadcountMetrics({ period: `${from},${to}` });
      const metrics = res.metrics;
      const hired = toFiniteNumber(metrics?.hiredThisPeriod);
      const departed = toFiniteNumber(metrics?.departedThisPeriod);
      return {
        activeStaff: toFiniteNumber(metrics?.current),
        momDelta: hired - departed,
        momDeltaPct: toFiniteNumber(metrics?.turnoverRate),
      };
    },
  });
};

export const useShiftMetrics = (params?: DateRangeParams): UseQueryResult<ShiftMetricsResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<ShiftMetricsResponse>({
    queryKey: ['insights-operations-shifts', { from, to }],
    queryFn: async (): Promise<ShiftMetricsResponse> => {
      const res = await operationClient.getShiftMetrics({ period: `${from},${to}` });
      const metrics = res.metrics;
      return {
        scheduled: toFiniteNumber(metrics?.totalShifts),
        completed: toFiniteNumber(metrics?.covered),
        noShows: toFiniteNumber(metrics?.missed),
        completionRatePct: toFiniteNumber(metrics?.coverageRate),
      };
    },
  });
};

export const useTimeOff = (params?: DateRangeParams): UseQueryResult<TimeOffResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<TimeOffResponse>({
    queryKey: ['insights-operations-time-off', { from, to }],
    queryFn: async (): Promise<TimeOffResponse> => {
      const res = await operationClient.getTimeOffMetrics({ period: `${from},${to}` });
      return {
        pending: toFiniteNumber(res.metrics?.pending),
        approved: toFiniteNumber(res.metrics?.approved),
      };
    },
  });
};

export const useUtilization = (params?: DateRangeParams): UseQueryResult<UtilizationResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<UtilizationResponse>({
    queryKey: ['insights-operations-utilization', { from, to }],
    queryFn: async (): Promise<UtilizationResponse> => {
      const res = await operationClient.getUtilization({ period: `${from},${to}` });
      const metrics = res.metrics;
      return {
        scheduledHours: toFiniteNumber(metrics?.totalHours),
        availableHours: toFiniteNumber(metrics?.availableHours),
        utilizationPct: toFiniteNumber(metrics?.avgUtilization),
      };
    },
  });
};
