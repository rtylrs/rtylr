import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { ProductServiceClient } from '@voxire/voxie/insights/product/v1/client';
import type {
  ProductKPIResponse,
  BestSellersResponse,
  ProductMarginsResponse,
  SlowMoversResponse,
  ProductCategoryMixResponse,
} from './types';

const productClient = createOneKitClient(ProductServiceClient, 'insights');

export const useProductKPIs = (): UseQueryResult<ProductKPIResponse> => {
  return useQuery<ProductKPIResponse>({
    queryKey: ['insights-products'],
    queryFn: async (): Promise<ProductKPIResponse> => {
      const res = await productClient.getProductPerformance({});
      return {
        totalSkus: toFiniteNumber(res.totalSkus),
        activeSkus: toFiniteNumber(res.activeSkus),
        outOfStock: toFiniteNumber(res.outOfStock),
        avgMarginPct: toFiniteNumber(res.avgMarginPct),
      };
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
      const res = await productClient.getBestSellers({ limit });
      return {
        products: (res.products ?? []).map((p) => ({
          productId: toFiniteNumber(p.productId),
          name: p.name ?? '',
          unitsSold: toFiniteNumber(p.unitsSold),
          revenueMinorUnits: toFiniteNumber(p.revenueMinorUnits),
        })),
      };
    },
  });
};

export const useProductMargins = (): UseQueryResult<ProductMarginsResponse> => {
  return useQuery<ProductMarginsResponse>({
    queryKey: ['insights-products-margins'],
    queryFn: async (): Promise<ProductMarginsResponse> => {
      const res = await productClient.getProductMargins({});
      return {
        buckets: (res.buckets ?? []).map((b) => ({
          label: b.label ?? '',
          lowerPct: toFiniteNumber(b.lowerPct),
          upperPct: toFiniteNumber(b.upperPct),
          productCount: toFiniteNumber(b.productCount),
        })),
      };
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
      const res = await productClient.getSlowMovers({ days });
      return {
        products: (res.products ?? []).map((p) => ({
          productId: toFiniteNumber(p.productId),
          name: p.name ?? '',
          stockQuantity: toFiniteNumber(p.stockQuantity),
        })),
      };
    },
  });
};

export const useProductCategoryMix = (): UseQueryResult<ProductCategoryMixResponse> => {
  return useQuery<ProductCategoryMixResponse>({
    queryKey: ['insights-products-category-mix'],
    queryFn: async (): Promise<ProductCategoryMixResponse> => {
      const res = await productClient.getCategoryPerformance({});
      return {
        categories: (res.categories ?? []).map((c) => ({
          categoryId: toFiniteNumber(c.categoryId),
          categoryName: c.categoryName ?? '',
          unitsSold: toFiniteNumber(c.unitsSold),
          revenueMinorUnits: toFiniteNumber(c.revenueMinorUnits),
        })),
      };
    },
  });
};
