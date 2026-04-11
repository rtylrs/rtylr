import { useQuery } from '@tanstack/react-query';
import { insightsService } from '../services/insightsService';

export function useProductPerformance() {
  return useQuery({
    queryKey: ['insights', 'products'],
    queryFn: () => insightsService.getProductPerformance().then((r) => r.data),
  });
}

export function useBestSellers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'best-sellers', limit],
    queryFn: () => insightsService.getBestSellers({ limit }).then((r) => r.data),
  });
}

export function useSlowMovers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'slow-movers', limit],
    queryFn: () => insightsService.getSlowMovers({ limit }).then((r) => r.data),
  });
}

export function useMarginAnalysis() {
  return useQuery({
    queryKey: ['insights', 'margins'],
    queryFn: () => insightsService.getMarginAnalysis().then((r) => r.data),
  });
}

export function useCategoryBreakdown() {
  return useQuery({
    queryKey: ['insights', 'categories'],
    queryFn: () => insightsService.getCategoryBreakdown().then((r) => r.data),
  });
}
