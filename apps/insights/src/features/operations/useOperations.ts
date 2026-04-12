import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';
import {
  operationsSummarySchema,
  staffUtilizationSchema,
  shiftCoverageSchema,
  timeOffTrendSchema,
  deptHeadcountSchema,
  hrDashboardSchema,
} from '../../shared/schemas/operations';

export function useOperationsSummary() {
  return useQuery({
    queryKey: ['insights', 'operations'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getOperationsSummary());
      return operationsSummarySchema.parse(res);
    },
  });
}

export function useStaffUtilization() {
  return useQuery({
    queryKey: ['insights', 'utilization'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getStaffUtilization());
      return staffUtilizationSchema.parse(res);
    },
  });
}

export function useShiftCoverage() {
  return useQuery({
    queryKey: ['insights', 'shifts'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getShiftCoverage());
      return shiftCoverageSchema.parse(res);
    },
  });
}

export function useTimeOffTrends() {
  return useQuery({
    queryKey: ['insights', 'time-off'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getTimeOffTrends());
      return timeOffTrendSchema.parse(res);
    },
  });
}

export function useDeptHeadcount() {
  return useQuery({
    queryKey: ['insights', 'headcount'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getDeptHeadcount());
      return deptHeadcountSchema.parse(res);
    },
  });
}

export function useHrDashboard() {
  return useQuery({
    queryKey: ['insights', 'hr-dashboard'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getHrDashboard());
      return hrDashboardSchema.parse(res);
    },
  });
}
