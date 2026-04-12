import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import {
  storeApi,
  CreateProductInput,
  UpdateProductInput,
  StoreProduct,
} from '../../shared/services/storeService';
import { storeProductSchema } from '../../shared/schemas/products';

export interface ProductFilters {
  search?: string;
  categoryId?: number;
  status?: StoreProduct['status'];
  limit?: number;
  offset?: number;
}

export function useProducts(filters?: ProductFilters) {
  return useQuery({
    queryKey: ['store-products', filters],
    queryFn: async () => {
      const res = await unwrap(
        storeApi.listProducts({
          search: filters?.search,
          categoryId: filters?.categoryId,
          status: filters?.status,
          limit: filters?.limit,
          offset: filters?.offset,
        })
      );
      const raw = (res as any)?.products ?? res;
      return z.array(storeProductSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useProduct(id: number | undefined) {
  return useQuery({
    queryKey: ['store-product', id],
    enabled: !!id,
    queryFn: async () => {
      const res = await unwrap(storeApi.getProduct(id!));
      return storeProductSchema.parse(res);
    },
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProductInput) => unwrap(storeApi.createProduct(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-products'] }),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateProductInput) => unwrap(storeApi.updateProduct(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-products'] }),
  });
}

export function useDeleteProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(storeApi.deleteProduct(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-products'] }),
  });
}
