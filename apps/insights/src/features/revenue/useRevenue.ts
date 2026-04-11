import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';

export function useRevenue(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue', period],
    queryFn: () => unwrap(insightsService.getRevenue({ period })),
  });
}

export function useRevenueByApp(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue-by-app', period],
    queryFn: () => unwrap(insightsService.getRevenueByApp({ period })),
  });
}

export function useRevenueComparison(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue-comparison', period],
    queryFn: () => unwrap(insightsService.getRevenueComparison({ period })),
  });
}
