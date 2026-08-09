import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { ProductServiceClient } from '@voxire/voxie/erp/product/v1/client';
import type {
  CreateProductRequest,
  CreateProductResponse,
  DeleteProductResponse,
  FullProduct,
  ListProductsResponse,
  Product,
  ProductAttribute,
  UpdateProductRequest,
  UpdateProductResponse,
} from '@voxire/voxie/erp/product/v1/types';

const productClient = createOneKitClient(ProductServiceClient, 'erp');

export type { FullProduct, Product, ProductAttribute };
export type CreateProductInput = CreateProductRequest;
export type UpdateProductInput = UpdateProductRequest;

export type ProductStockStatus = 'In Stock' | 'Low Stock' | 'Out of Stock';

export interface ProductFilters {
  search?: string;
  categoryId?: string;
  status?: ProductStockStatus;
  sort?: 'name' | 'stock' | 'price' | 'category';
  sortDir?: 'asc' | 'desc';
  page?: number;
  pageSize?: number;
}

export function getProductStockStatus(
  product: Pick<Product, 'stockQuantity' | 'lowStockThreshold'>
): ProductStockStatus {
  if ((product.stockQuantity ?? 0) <= 0) return 'Out of Stock';
  if ((product.stockQuantity ?? 0) <= (product.lowStockThreshold ?? 0)) return 'Low Stock';
  return 'In Stock';
}

function pageOffset(page: number, pageSize: number): number {
  return Math.max(page - 1, 0) * pageSize;
}

export function useProducts(filters?: ProductFilters) {
  const page = filters?.page ?? 1;
  const pageSize = filters?.pageSize ?? 50;
  const search = filters?.search?.trim() ?? '';
  const categoryId = filters?.categoryId ?? '0';

  return useQuery<ListProductsResponse>({
    queryKey: ['products', { ...filters, search, categoryId, page, pageSize }],
    queryFn: () =>
      productClient.listProducts({
        search,
        categoryId,
        limit: pageSize,
        offset: pageOffset(page, pageSize),
      }),
  });
}

export function useProduct(id: string | undefined) {
  return useQuery<Product | null>({
    queryKey: ['product', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      const response = await productClient.getProduct({ id });
      return response.product ?? null;
    },
  });
}

export function useProductFull(id: string | undefined) {
  return useQuery<FullProduct | null>({
    queryKey: ['product-full', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      const response = await productClient.getProductFull({ id });
      return response.product ?? null;
    },
  });
}

export function useProductAttributes(id: string | undefined) {
  return useQuery<ProductAttribute[]>({
    queryKey: ['product-attributes', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return [];
      const response = await productClient.listProductAttributes({ id });
      return response.attributes ?? [];
    },
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation<CreateProductResponse, Error, CreateProductInput>({
    mutationFn: (data) => productClient.createProduct(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
    onError: (error) =>
      toast.error(error instanceof Error ? error.message : 'Failed to create product'),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation<UpdateProductResponse, Error, UpdateProductInput>({
    mutationFn: (data) => productClient.updateProduct(data),
    onSuccess: (_res, variables) => {
      qc.invalidateQueries({ queryKey: ['products'] });
      qc.invalidateQueries({ queryKey: ['product', variables.id] });
      qc.invalidateQueries({ queryKey: ['product-full', variables.id] });
    },
    onError: () => toast.error('Failed to update product'),
  });
}

export function useDeleteProduct() {
  const qc = useQueryClient();
  return useMutation<DeleteProductResponse, Error, string>({
    mutationFn: (id) => productClient.deleteProduct({ id }),
    onSuccess: (_res, id) => {
      qc.invalidateQueries({ queryKey: ['products'] });
      qc.invalidateQueries({ queryKey: ['product', id] });
      qc.invalidateQueries({ queryKey: ['product-full', id] });
    },
    onError: () => toast.error('Failed to delete product'),
  });
}
