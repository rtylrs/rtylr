import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  Product,
  ListProductsResponse,
  CreateProductRequest,
  CreateProductResponse,
  UpdateProductRequest,
  UpdateProductResponse,
  DeleteProductResponse,
} from '@voxire/contracts/erp/v1/products';
import type { UseQueryResult, UseMutationResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface ListProductsParams {
  search?: string;
  categoryId?: number;
  limit?: number;
  offset?: number;
}

export const useListProducts = (params: ListProductsParams = {}): UseQueryResult<Product[]> => {
  return useQuery<Product[]>({
    queryKey: ['erp-products', 'filter', params],
    queryFn: async (): Promise<Product[]> => {
      const res = await unwrap(safe.get<ListProductsResponse>('/erp/products', { params }));
      return (res as ListProductsResponse).products ?? [];
    },
  });
};

export const useGetProduct = (id: number | undefined): UseQueryResult<Product | null> => {
  return useQuery<Product | null>({
    queryKey: ['erp-product', id],
    enabled: !!id,
    queryFn: async (): Promise<Product | null> => {
      if (!id) return null;
      const res = await unwrap(safe.get<{ product: Product } | Product>(`/erp/products/${id}`));
      const maybeEnvelope = res as { product?: Product };
      if (maybeEnvelope && typeof maybeEnvelope === 'object' && 'product' in maybeEnvelope) {
        return maybeEnvelope.product ?? null;
      }
      return res as Product;
    },
  });
};

export const useCreateProduct = (): UseMutationResult<
  CreateProductResponse,
  Error,
  CreateProductRequest
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateProductRequest): Promise<CreateProductResponse> => {
      const res = await unwrap(safe.post<CreateProductResponse>('/erp/products', data));
      return res as CreateProductResponse;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-products'] });
      toast.success(t('erp.productCreated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateProduct = (): UseMutationResult<
  UpdateProductResponse,
  Error,
  UpdateProductRequest
> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpdateProductRequest): Promise<UpdateProductResponse> => {
      const res = await unwrap(safe.put<UpdateProductResponse>(`/erp/products/${data.id}`, data));
      return res as UpdateProductResponse;
    },
    onSuccess: (_res: UpdateProductResponse, variables: UpdateProductRequest): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-products'] });
      queryClient.invalidateQueries({ queryKey: ['erp-product', variables.id] });
      toast.success(t('erp.productUpdated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteProduct = (): UseMutationResult<DeleteProductResponse, Error, number> => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (id: number): Promise<DeleteProductResponse> => {
      const res = await unwrap(safe.delete<DeleteProductResponse>(`/erp/products/${id}`));
      return res as DeleteProductResponse;
    },
    onSuccess: (): void => {
      queryClient.invalidateQueries({ queryKey: ['erp-products'] });
      toast.success(t('erp.productDeleted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useBarcodeLookup = (code: string): UseQueryResult<Product | null> => {
  return useQuery<Product | null>({
    queryKey: ['erp-product', 'barcode', code],
    enabled: false,
    queryFn: async (): Promise<Product | null> => {
      if (!code) return null;
      const res = await unwrap(
        safe.get<{ product: Product } | Product>(
          `/erp/products/barcode/${encodeURIComponent(code)}`
        )
      );
      const maybeEnvelope = res as { product?: Product };
      if (maybeEnvelope && typeof maybeEnvelope === 'object' && 'product' in maybeEnvelope) {
        return maybeEnvelope.product ?? null;
      }
      return res as Product;
    },
  });
};
