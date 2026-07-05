import { createOneKitClient } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { OperationServiceClient } from '@voxire/voxie/insights/operation/v1/service_client';
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
        activeStaff: Number(res.activeStaff),
        shiftsThisWeek: Number(res.shiftsThisWeek),
        pendingTimeOff: Number(res.pendingTimeOff),
        openLeaveRequests: Number(res.openLeaveRequests),
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
      const hired = metrics?.hiredThisPeriod ?? 0;
      const departed = metrics?.departedThisPeriod ?? 0;
      return {
        activeStaff: metrics?.current ?? 0,
        momDelta: hired - departed,
        momDeltaPct: metrics?.turnoverRate ?? 0,
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
        scheduled: metrics?.totalShifts ?? 0,
        completed: metrics?.covered ?? 0,
        noShows: metrics?.missed ?? 0,
        completionRatePct: metrics?.coverageRate ?? 0,
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
        pending: res.metrics?.pending ?? 0,
        approved: res.metrics?.approved ?? 0,
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
        scheduledHours: metrics?.totalHours ?? 0,
        availableHours: metrics?.availableHours ?? 0,
        utilizationPct: metrics?.avgUtilization ?? 0,
      };
    },
  });
};
