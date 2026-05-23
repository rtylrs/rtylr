import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  erpApi,
  CreateProductInput,
  UpdateProductInput,
  Product,
} from '../../shared/services/erpApi';
import { productSchema } from '../../shared/schemas/products';
import { unwrapList } from '../../shared/lib/unwrapList';

export interface ProductFilters {
  search?: string;
  categoryId?: number;
  category?: string;
  brand?: string;
  status?: Product['status'];
  sort?: 'name' | 'stock' | 'price' | 'category';
  sortDir?: 'asc' | 'desc';
  page?: number;
  pageSize?: number;
}

export interface ProductListResult {
  products: Product[];
  hasMore: boolean;
  page: number;
  pageSize: number;
}

function applyFilters(list: Product[], f?: ProductFilters): Product[] {
  if (!f) return list;
  let out = list.slice();
  if (f.search) {
    const q = f.search.toLowerCase();
    out = out.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.sku.toLowerCase().includes(q) ||
        p.category.toLowerCase().includes(q) ||
        (p.brand && p.brand.toLowerCase().includes(q))
    );
  }
  if (f.category) out = out.filter((p) => p.category === f.category);
  if (f.brand) out = out.filter((p) => p.brand === f.brand);
  if (f.status) out = out.filter((p) => p.status === f.status);
  if (f.sort) {
    const dir = f.sortDir === 'desc' ? -1 : 1;
    out.sort((a, b) => {
      if (f.sort === 'name') return a.name.localeCompare(b.name) * dir;
      if (f.sort === 'stock') return (a.stock - b.stock) * dir;
      if (f.sort === 'price') return (a.salePrice - b.salePrice) * dir;
      if (f.sort === 'category') return a.category.localeCompare(b.category) * dir;
      return 0;
    });
  }
  return out;
}

export function useProducts(filters?: ProductFilters) {
  const page = filters?.page ?? 1;
  const pageSize = filters?.pageSize ?? 50;
  return useQuery<ProductListResult>({
    queryKey: ['products', { ...filters, page, pageSize }],
    queryFn: async () => {
      const res = await unwrap(
        erpApi.listProducts({
          search: filters?.search,
          category_id: filters?.categoryId,
          page,
          page_size: pageSize,
        })
      );
      const parsed = z.array(productSchema).parse(unwrapList(res, 'products')) as Product[];
      const products = applyFilters(parsed, filters);
      const hasMore = (res as unknown as Record<string, unknown>)?.has_more === true;
      return { products, hasMore, page, pageSize };
    },
  });
}

export function useProduct(id: number | undefined) {
  return useQuery<Product | null>({
    queryKey: ['product', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      const res = await unwrap(erpApi.getProduct(id));
      return productSchema.parse(res) as Product;
    },
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProductInput) => unwrap(erpApi.createProduct(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
    onError: () => toast.error('Failed to create product'),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateProductInput) => unwrap(erpApi.updateProduct(data)),
    onSuccess: (_res, variables) => {
      // Invalidate both the list cache and the individual product cache
      // so the detail screen sees the edited name on its next render.
      qc.invalidateQueries({ queryKey: ['products'] });
      qc.invalidateQueries({ queryKey: ['product', variables.id] });
    },
    onError: () => toast.error('Failed to update product'),
  });
}

export function useDeleteProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.deleteProduct(id)),
    onSuccess: (_res, id) => {
      qc.invalidateQueries({ queryKey: ['products'] });
      qc.invalidateQueries({ queryKey: ['product', id] });
    },
    onError: () => toast.error('Failed to delete product'),
  });
}
