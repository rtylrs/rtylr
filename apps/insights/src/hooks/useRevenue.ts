import { useQuery } from '@tanstack/react-query';
import { insightsService } from '../services/insightsService';

export function useRevenue(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue', period],
    queryFn: () => insightsService.getRevenue({ period }).then((r) => r.data),
  });
}

export function useRevenueByApp(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue-by-app', period],
    queryFn: () => insightsService.getRevenueByApp({ period }).then((r) => r.data),
  });
}

export function useRevenueComparison(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue-comparison', period],
    queryFn: () => insightsService.getRevenueComparison({ period }).then((r) => r.data),
  });
}
