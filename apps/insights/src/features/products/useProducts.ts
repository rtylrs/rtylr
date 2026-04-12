import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';
import {
  productPerformanceSchema,
  productRankingSchema,
  marginAnalysisSchema,
  categoryBreakdownSchema,
} from '../../shared/schemas/products';

export function useProductPerformance() {
  return useQuery({
    queryKey: ['insights', 'products'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getProductPerformance());
      return productPerformanceSchema.parse(res);
    },
  });
}

export function useBestSellers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'best-sellers', limit],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getBestSellers({ limit }));
      const raw = (res as any)?.products ?? res;
      return z.array(productRankingSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useSlowMovers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'slow-movers', limit],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getSlowMovers({ limit }));
      const raw = (res as any)?.products ?? res;
      return z.array(productRankingSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useMarginAnalysis() {
  return useQuery({
    queryKey: ['insights', 'margins'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getMarginAnalysis());
      return marginAnalysisSchema.parse(res);
    },
  });
}

export function useCategoryBreakdown() {
  return useQuery({
    queryKey: ['insights', 'categories'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getCategoryBreakdown());
      const raw = (res as any)?.categories ?? res;
      return z.array(categoryBreakdownSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}
