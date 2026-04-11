import { useQuery } from '@tanstack/react-query';
import { insightsService } from '../services/insightsService';

export function useOperationsSummary() {
  return useQuery({
    queryKey: ['insights', 'operations'],
    queryFn: () => insightsService.getOperationsSummary().then((r) => r.data),
  });
}

export function useStaffUtilization() {
  return useQuery({
    queryKey: ['insights', 'utilization'],
    queryFn: () => insightsService.getStaffUtilization().then((r) => r.data),
  });
}

export function useShiftCoverage() {
  return useQuery({
    queryKey: ['insights', 'shifts'],
    queryFn: () => insightsService.getShiftCoverage().then((r) => r.data),
  });
}

export function useTimeOffTrends() {
  return useQuery({
    queryKey: ['insights', 'time-off'],
    queryFn: () => insightsService.getTimeOffTrends().then((r) => r.data),
  });
}

export function useDeptHeadcount() {
  return useQuery({
    queryKey: ['insights', 'headcount'],
    queryFn: () => insightsService.getDeptHeadcount().then((r) => r.data),
  });
}

export function useHrDashboard() {
  return useQuery({
    queryKey: ['insights', 'hr-dashboard'],
    queryFn: () => insightsService.getHrDashboard().then((r) => r.data),
  });
}
