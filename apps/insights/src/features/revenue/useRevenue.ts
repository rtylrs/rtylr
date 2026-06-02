import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';
import {
  revenueDataSchema,
  revenueByAppSchema,
  revenueComparisonSchema,
} from '../../shared/schemas/revenue';

export function useRevenue(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue', period],
    queryFn: async () => {
      const res = await unwrap(insightsService.getRevenue({ period }));
      return revenueDataSchema.parse(res);
    },
  });
}

export function useRevenueByApp(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue-by-app', period],
    queryFn: async () => {
      const res = await unwrap(insightsService.getRevenueByApp({ period }));
      const raw = (res as Record<string, unknown>)?.byApp;
      return z.array(revenueByAppSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useRevenueComparison(period?: string) {
  return useQuery({
    queryKey: ['insights', 'revenue-comparison', period],
    queryFn: async () => {
      const res = await unwrap(insightsService.getRevenueComparison({ period }));
      return revenueComparisonSchema.parse(res);
    },
  });
}
