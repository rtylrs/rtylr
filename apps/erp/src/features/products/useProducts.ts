import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { validateContractPayload } from '@voxire/contracts/validation/metadata';
import {
  erpApi,
  CreateProductInput,
  UpdateProductInput,
  Product,
} from '../../shared/services/erpApi';
import { productSchema } from '../../shared/schemas/products';
import { unwrapList } from '../../shared/lib/unwrapList';

function assertContractPayload(messageName: string, payload: Record<string, unknown>): void {
  const issues = validateContractPayload(messageName, payload);
  if (issues.length > 0) {
    throw new Error(issues[0].message);
  }
}

export interface ProductFilters {
  search?: string;
  categoryId?: number;
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

// The backend (erp/v1 Product) returns categoryId/stockQuantity/lowStockThreshold
// and has no `category` string, `stock`, `lowStock`, `status`, or `brand` field.
// The shared productSchema/Product type still uses the legacy display shape, so we
// normalise the real backend payload into that shape before parsing. Status is a
// display-only value computed from stock vs. threshold — the backend never sends it.
interface BackendProduct {
  id: number;
  name?: string | null;
  sku?: string | null;
  categoryId?: number;
  stockQuantity?: number;
  lowStockThreshold?: number;
  reorderQuantity?: number;
  costPrice?: number;
  salePrice?: number;
  barcode?: string;
  description?: string;
  unit?: string;
  image?: string | null;
  imageUrl?: string | null;
  image_url?: string | null;
}

function computeStatus(stock: number, lowStock: number): Product['status'] {
  if (stock <= 0) return 'Out of Stock';
  if (stock <= lowStock) return 'Low Stock';
  return 'In Stock';
}

function normalizeProduct(raw: BackendProduct): Record<string, unknown> {
  const stock = raw.stockQuantity ?? 0;
  const lowStock = raw.lowStockThreshold ?? 0;
  const imageUrl = raw.imageUrl ?? raw.image_url ?? raw.image ?? null;
  return {
    id: raw.id,
    name: raw.name?.trim() ?? '',
    sku: raw.sku?.trim() ?? '',
    // Backend has no category name — only categoryId. Keep the id around for
    // filtering and leave the display string empty so the schema doesn't throw.
    category: '',
    categoryId: raw.categoryId,
    stock,
    lowStock,
    costPrice: raw.costPrice ?? 0,
    salePrice: raw.salePrice ?? 0,
    status: computeStatus(stock, lowStock),
    barcode: raw.barcode ?? null,
    description: raw.description ?? null,
    imageUrl,
    unit: raw.unit ?? null,
    reorderPoint: raw.lowStockThreshold ?? null,
  };
}

// Internal product shape including categoryId so we can filter by category.
type ProductWithCategoryId = Product & { categoryId?: number };

function applyFilters(list: ProductWithCategoryId[], f?: ProductFilters): Product[] {
  let out = list.filter((p) => p.name.trim() !== '');
  if (!f) return out;
  if (f.search) {
    const q = f.search.toLowerCase();
    out = out.filter(
      (p) =>
        p.name.toLowerCase().includes(q) ||
        p.sku.toLowerCase().includes(q) ||
        (p.brand && p.brand.toLowerCase().includes(q))
    );
  }
  if (f.categoryId !== undefined) out = out.filter((p) => p.categoryId === f.categoryId);
  if (f.brand) out = out.filter((p) => p.brand === f.brand);
  if (f.status) out = out.filter((p) => p.status === f.status);
  if (f.sort) {
    const dir = f.sortDir === 'desc' ? -1 : 1;
    out.sort((a, b) => {
      if (f.sort === 'name') return a.name.localeCompare(b.name) * dir;
      if (f.sort === 'stock') return (a.stock - b.stock) * dir;
      if (f.sort === 'price') return (a.salePrice - b.salePrice) * dir;
      if (f.sort === 'category') return ((a.categoryId ?? 0) - (b.categoryId ?? 0)) * dir;
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
      const raw = unwrapList<BackendProduct>(res, 'products');
      const parsed: ProductWithCategoryId[] = raw.map((r) => {
        const normalized = normalizeProduct(r);
        return {
          ...(productSchema.parse(normalized) as Product),
          categoryId: r.categoryId,
        };
      });
      const products = applyFilters(parsed, filters);
      // @voxire/api converts has_more -> hasMore on the way in.
      const hasMore = (res as unknown as Record<string, unknown>)?.hasMore === true;
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
      // GET /erp/products/{id} returns { product: {...} }. Some callers/tests may
      // hand back the bare product, so accept either shape.
      const resObj = res as unknown as Record<string, unknown> | null;
      const rawProduct =
        resObj && typeof resObj === 'object' && 'product' in resObj
          ? (resObj.product as BackendProduct)
          : (res as unknown as BackendProduct);
      const parsed = productSchema.parse(normalizeProduct(rawProduct)) as Product;
      return { ...parsed, categoryId: rawProduct.categoryId } as Product;
    },
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProductInput) => {
      assertContractPayload('erp.product.v1.CreateProductRequest', {
        ...(data as unknown as Record<string, unknown>),
      });
      return unwrap(erpApi.createProduct(data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
    onError: (error) =>
      toast.error(error instanceof Error ? error.message : 'Failed to create product'),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateProductInput) => {
      assertContractPayload('erp.product.v1.UpdateProductRequest', {
        ...(data as unknown as Record<string, unknown>),
      });
      return unwrap(erpApi.updateProduct(data));
    },
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
