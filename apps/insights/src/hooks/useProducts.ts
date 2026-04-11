import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../services/insightsService';

export function useProductPerformance() {
  return useQuery({
    queryKey: ['insights', 'products'],
    queryFn: () => unwrap(insightsService.getProductPerformance()),
  });
}

export function useBestSellers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'best-sellers', limit],
    queryFn: () => unwrap(insightsService.getBestSellers({ limit })),
  });
}

export function useSlowMovers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'slow-movers', limit],
    queryFn: () => unwrap(insightsService.getSlowMovers({ limit })),
  });
}

export function useMarginAnalysis() {
  return useQuery({
    queryKey: ['insights', 'margins'],
    queryFn: () => unwrap(insightsService.getMarginAnalysis()),
  });
}

export function useCategoryBreakdown() {
  return useQuery({
    queryKey: ['insights', 'categories'],
    queryFn: () => unwrap(insightsService.getCategoryBreakdown()),
  });
}
