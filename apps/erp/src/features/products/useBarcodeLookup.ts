import { unwrap } from '@voxire/api';
import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, Product } from '../../shared/services/erpApi';
import { productSchema } from '../../shared/schemas/products';

// GET /erp/products/barcode/{code} returns { product: {...} } using the erp/v1
// Product shape (categoryId/stockQuantity/lowStockThreshold, no category string
// or status). Normalise it into the shared Product display shape, mirroring
// useProducts. Returns null when no product matches the barcode.
interface BackendProduct {
  id: number;
  name: string;
  sku: string;
  categoryId?: number;
  stockQuantity?: number;
  lowStockThreshold?: number;
  costPrice?: number;
  salePrice?: number;
  barcode?: string;
  description?: string;
  unit?: string;
}

function computeStatus(stock: number, lowStock: number): Product['status'] {
  if (stock <= 0) return 'Out of Stock';
  if (stock <= lowStock) return 'Low Stock';
  return 'In Stock';
}

function normalizeProduct(raw: BackendProduct): Record<string, unknown> {
  const stock = raw.stockQuantity ?? 0;
  const lowStock = raw.lowStockThreshold ?? 0;
  return {
    id: raw.id,
    name: raw.name,
    sku: raw.sku,
    category: '',
    stock,
    lowStock,
    costPrice: raw.costPrice ?? 0,
    salePrice: raw.salePrice ?? 0,
    status: computeStatus(stock, lowStock),
    barcode: raw.barcode ?? null,
    description: raw.description ?? null,
    unit: raw.unit ?? null,
    reorderPoint: raw.lowStockThreshold ?? null,
  };
}

export function useBarcodeLookup() {
  return useMutation<Product | null, unknown, string>({
    mutationFn: async (barcode: string) => {
      const res = await unwrap(erpApi.lookupBarcode(barcode));
      if (!res || typeof res !== 'object') return null;
      const resObj = res as unknown as Record<string, unknown>;
      const rawProduct =
        'product' in resObj
          ? (resObj.product as BackendProduct | null)
          : (res as unknown as BackendProduct);
      if (!rawProduct) return null;
      return productSchema.parse(normalizeProduct(rawProduct)) as Product;
    },
    onError: () => toast.error('Failed to look up barcode'),
  });
}
