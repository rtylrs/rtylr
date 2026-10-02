import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
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
import { invalidateStockDerivedQueries } from '../../shared/lib/stockInvalidation';

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

const EXPORT_PAGE_SIZE = 100;

export async function fetchAllProducts(filters: ProductFilters = {}): Promise<Product[]> {
  const products: Product[] = [];
  let offset = 0;

  while (true) {
    const response = await productClient.listProducts({
      search: filters.search?.trim() ?? '',
      categoryId: filters.categoryId ?? '0',
      limit: EXPORT_PAGE_SIZE,
      offset,
    });
    const page = response.products ?? [];
    products.push(...page);

    const hasMore =
      response.hasMore ??
      (response.total !== undefined
        ? products.length < response.total
        : page.length === EXPORT_PAGE_SIZE);
    if (!hasMore || page.length === 0) return filterAndSortProducts(products, filters);
    offset += page.length;
  }
}

export function getProductStockStatus(
  product: Pick<Product, 'stockQuantity' | 'lowStockThreshold'>
): ProductStockStatus {
  if ((product.stockQuantity ?? 0) <= 0) return 'Out of Stock';
  if ((product.stockQuantity ?? 0) <= (product.lowStockThreshold ?? 0)) return 'Low Stock';
  return 'In Stock';
}

function filterAndSortProducts(rawProducts: Product[], filters: ProductFilters): Product[] {
  let products = rawProducts;

  if (filters.status) {
    products = products.filter((product) => getProductStockStatus(product) === filters.status);
  }

  if (filters.sort) {
    const direction = filters.sortDir === 'desc' ? -1 : 1;
    products = [...products].sort((a, b) => {
      if (filters.sort === 'name') return (a.name ?? '').localeCompare(b.name ?? '') * direction;
      if (filters.sort === 'stock') {
        return ((a.stockQuantity ?? 0) - (b.stockQuantity ?? 0)) * direction;
      }
      if (filters.sort === 'price') return ((a.salePrice ?? 0) - (b.salePrice ?? 0)) * direction;
      if (filters.sort === 'category') {
        return (a.categoryId ?? '').localeCompare(b.categoryId ?? '') * direction;
      }
      return 0;
    });
  }

  return products;
}

function pageOffset(page: number, pageSize: number): number {
  return Math.max(page - 1, 0) * pageSize;
}

export function useProducts(filters?: ProductFilters) {
  const page = filters?.page ?? 1;
  const pageSize = filters?.pageSize ?? 50;
  const search = filters?.search?.trim() ?? '';
  const categoryId = filters?.categoryId ?? '0';
  const status = filters?.status;
  const sort = filters?.sort;
  const sortDir = filters?.sortDir;
  const needsClientFiltering = status !== undefined || sort !== undefined;
  const offset = pageOffset(page, pageSize);

  const query = useQuery<ListProductsResponse>({
    queryKey: needsClientFiltering
      ? ['products', { search, categoryId, status, sort, sortDir }]
      : ['products', { search, categoryId, page, pageSize }],
    queryFn: async () => {
      if (needsClientFiltering) {
        const products = await fetchAllProducts({ search, categoryId, status, sort, sortDir });
        return { products, total: products.length };
      }

      return productClient.listProducts({
        search,
        categoryId,
        limit: pageSize,
        offset,
      });
    },
  });

  if (!needsClientFiltering || !query.data) return query;

  const allProducts = query.data.products ?? [];
  const products = allProducts.slice(offset, offset + pageSize);
  return {
    ...query,
    data: {
      ...query.data,
      products,
      page,
      pageSize,
      total: allProducts.length,
      hasMore: offset + products.length < allProducts.length,
    },
  };
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
    onSuccess: () => invalidateStockDerivedQueries(qc),
    onError: (error) =>
      toast.error(error instanceof Error ? error.message : 'Failed to create product'),
  });
}

export function useUpdateProduct() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<UpdateProductResponse, Error, UpdateProductInput>({
    mutationFn: (data) => productClient.updateProduct(data),
    onSuccess: (_res, variables) => {
      invalidateStockDerivedQueries(qc);
      qc.invalidateQueries({ queryKey: ['product', variables.id] });
      qc.invalidateQueries({ queryKey: ['product-full', variables.id] });
    },
    onError: () => toast.error(t('erp.products.updateFailed')),
  });
}

export function useDeleteProduct() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<DeleteProductResponse, Error, string>({
    mutationFn: (id) => productClient.deleteProduct({ id }),
    onSuccess: (_res, id) => {
      invalidateStockDerivedQueries(qc);
      qc.invalidateQueries({ queryKey: ['product', id] });
      qc.invalidateQueries({ queryKey: ['product-full', id] });
    },
    onError: () => toast.error(t('erp.products.deleteFailed')),
  });
}
