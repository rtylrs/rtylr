import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi, CreateProductInput, UpdateProductInput, Product } from '../../shared/services/erpApi';

export interface ProductFilters {
  search?: string;
  categoryId?: number;
  category?: string;
  brand?: string;
  status?: Product['status'];
  sort?: 'name' | 'stock' | 'price' | 'category';
  sortDir?: 'asc' | 'desc';
  limit?: number;
  offset?: number;
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
  return useQuery({
    queryKey: ['products', filters],
    queryFn: async () => {
      const data = await unwrap(erpApi.listProducts({
        search: filters?.search,
        categoryId: filters?.categoryId,
        limit: filters?.limit,
        offset: filters?.offset,
      }));
      return applyFilters(data ?? [], filters);
    },
  });
}

export function useProduct(id: number | undefined) {
  return useQuery<Product | null>({
    queryKey: ['product', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      return unwrap(erpApi.getProduct(id));
    },
  });
}

export function useCreateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateProductInput) => unwrap(erpApi.createProduct(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  });
}

export function useUpdateProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateProductInput) => unwrap(erpApi.updateProduct(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  });
}

export function useDeleteProduct() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.deleteProduct(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['products'] }),
  });
}
