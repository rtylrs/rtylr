import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { storeApi, CreateProductInput, UpdateProductInput, StoreProduct } from '../services/storeService';

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
      const r = await storeApi.listProducts({
        search: filters?.search,
        categoryId: filters?.categoryId,
        status: filters?.status,
        limit: filters?.limit,
        offset: filters?.offset,
      });
      return r.data ?? [];
    },
  });
}

export function useProduct(id: number | undefined) {
  return useQuery<StoreProduct | null>({
    queryKey: ['store-product', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      const r = await storeApi.getProduct(id);
      return r.data ?? null;
    },
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProductInput) => storeApi.createProduct(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-products'] }),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateProductInput) => storeApi.updateProduct(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-products'] }),
  });
}

export function useDeleteProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => storeApi.deleteProduct(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-products'] }),
  });
}
