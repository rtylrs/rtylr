import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
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
    queryFn: () => unwrap(storeApi.listProducts({
      search: filters?.search,
      categoryId: filters?.categoryId,
      status: filters?.status,
      limit: filters?.limit,
      offset: filters?.offset,
    })),
  });
}

export function useProduct(id: number | undefined) {
  return useQuery<StoreProduct | null>({
    queryKey: ['store-product', id],
    enabled: !!id,
    queryFn: () => unwrap(storeApi.getProduct(id!)),
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
