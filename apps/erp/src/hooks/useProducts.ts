import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi, CreateProductInput, UpdateProductInput } from '../services/erpApi';

export function useProducts(params?: { search?: string; categoryId?: number; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['products', params],
    queryFn: () => erpApi.listProducts(params).then((r) => r.data || []),
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProductInput) => erpApi.createProduct(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateProductInput) => erpApi.updateProduct(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  });
}

export function useDeleteProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => erpApi.deleteProduct(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  });
}
