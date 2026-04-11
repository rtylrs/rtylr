import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../services/insightsService';

export function useOperationsSummary() {
  return useQuery({
    queryKey: ['insights', 'operations'],
    queryFn: () => unwrap(insightsService.getOperationsSummary()),
  });
}

export function useStaffUtilization() {
  return useQuery({
    queryKey: ['insights', 'utilization'],
    queryFn: () => unwrap(insightsService.getStaffUtilization()),
  });
}

export function useShiftCoverage() {
  return useQuery({
    queryKey: ['insights', 'shifts'],
    queryFn: () => unwrap(insightsService.getShiftCoverage()),
  });
}

export function useTimeOffTrends() {
  return useQuery({
    queryKey: ['insights', 'time-off'],
    queryFn: () => unwrap(insightsService.getTimeOffTrends()),
  });
}

export function useDeptHeadcount() {
  return useQuery({
    queryKey: ['insights', 'headcount'],
    queryFn: () => unwrap(insightsService.getDeptHeadcount()),
  });
}

export function useHrDashboard() {
  return useQuery({
    queryKey: ['insights', 'hr-dashboard'],
    queryFn: () => unwrap(insightsService.getHrDashboard()),
  });
}
