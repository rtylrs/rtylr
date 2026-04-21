import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import type {
  ProductKPIResponse,
  BestSellersResponse,
  ProductMarginsResponse,
  SlowMoversResponse,
  ProductCategoryMixResponse,
} from './types';

export const useProductKPIs = (): UseQueryResult<ProductKPIResponse> => {
  return useQuery<ProductKPIResponse>({
    queryKey: ['insights-products'],
    queryFn: async (): Promise<ProductKPIResponse> => {
      const res = await unwrap(safe.get<ProductKPIResponse>('/insights/products'));
      return res as ProductKPIResponse;
    },
  });
};

interface BestSellersParams {
  limit?: number;
}

export const useBestSellers = (params?: BestSellersParams): UseQueryResult<BestSellersResponse> => {
  const limit = params?.limit ?? 10;
  return useQuery<BestSellersResponse>({
    queryKey: ['insights-products-best-sellers', { limit }],
    queryFn: async (): Promise<BestSellersResponse> => {
      const res = await unwrap(
        safe.get<BestSellersResponse>('/insights/products/best-sellers', { params: { limit } })
      );
      return res as BestSellersResponse;
    },
  });
};

export const useProductMargins = (): UseQueryResult<ProductMarginsResponse> => {
  return useQuery<ProductMarginsResponse>({
    queryKey: ['insights-products-margins'],
    queryFn: async (): Promise<ProductMarginsResponse> => {
      const res = await unwrap(safe.get<ProductMarginsResponse>('/insights/products/margins'));
      return res as ProductMarginsResponse;
    },
  });
};

interface SlowMoversParams {
  days?: number;
}

export const useSlowMovers = (params?: SlowMoversParams): UseQueryResult<SlowMoversResponse> => {
  const days = params?.days ?? 60;
  return useQuery<SlowMoversResponse>({
    queryKey: ['insights-products-slow-movers', { days }],
    queryFn: async (): Promise<SlowMoversResponse> => {
      const res = await unwrap(
        safe.get<SlowMoversResponse>('/insights/products/slow-movers', { params: { days } })
      );
      return res as SlowMoversResponse;
    },
  });
};

export const useProductCategoryMix = (): UseQueryResult<ProductCategoryMixResponse> => {
  return useQuery<ProductCategoryMixResponse>({
    queryKey: ['insights-products-category-mix'],
    queryFn: async (): Promise<ProductCategoryMixResponse> => {
      const res = await unwrap(
        safe.get<ProductCategoryMixResponse>('/insights/products/categories')
      );
      return res as ProductCategoryMixResponse;
    },
  });
};
