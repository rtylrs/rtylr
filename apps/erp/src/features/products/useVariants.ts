import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, ProductVariant } from '../../shared/services/erpApi';
import { productVariantSchema } from '../../shared/schemas/products';
import { unwrapList } from '../../shared/lib/unwrapList';

// The backend (variants module) returns stockQuantity (not stock), salePrice
// (not price), and attributes as a JSON STRING (not an object). The shared
// productVariantSchema/ProductVariant type still uses the legacy shape, so we
// normalise the real payload into that shape before parsing.
interface BackendVariant {
  id: number;
  productId: number;
  name: string;
  sku: string;
  barcode?: string;
  costPrice?: number;
  salePrice?: number;
  stockQuantity?: number;
  attributes?: string;
  isActive?: boolean;
}

function parseAttributes(raw: string | undefined): Record<string, string> {
  if (!raw) return {};
  try {
    const parsed: unknown = JSON.parse(raw);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      const out: Record<string, string> = {};
      for (const [k, v] of Object.entries(parsed as Record<string, unknown>)) {
        out[k] = String(v);
      }
      return out;
    }
  } catch {
    // Malformed JSON in the DB — degrade to no attributes rather than throwing.
  }
  return {};
}

function normalizeVariant(raw: BackendVariant): Record<string, unknown> {
  return {
    id: raw.id,
    productId: raw.productId,
    name: raw.name,
    sku: raw.sku,
    attributes: parseAttributes(raw.attributes),
    stock: raw.stockQuantity ?? 0,
    price: raw.salePrice ?? 0,
  };
}

export function useVariants(productId?: number) {
  return useQuery<ProductVariant[]>({
    queryKey: ['variants', productId],
    queryFn: async () => {
      const res = await unwrap(erpApi.listVariants(productId));
      const raw = unwrapList<BackendVariant>(res, 'variants');
      return raw.map((r) => productVariantSchema.parse(normalizeVariant(r)) as ProductVariant);
    },
  });
}

export function useCreateVariant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ProductVariant>) => {
      // Map the legacy display shape to the backend CreateVariantRequest:
      // stock -> stock_quantity, price -> sale_price, and attributes must be a
      // JSON STRING (the backend stores it as a string column). @voxire/api will
      // snake_case the request body keys on the way out.
      const body = {
        productId: data.productId,
        name: data.name,
        sku: data.sku,
        salePrice: data.price ?? 0,
        stockQuantity: data.stock ?? 0,
        attributes: JSON.stringify(data.attributes ?? {}),
      };
      // The backend stores attributes as a JSON string, so the request body
      // intentionally diverges from the ProductVariant display type here.
      return unwrap(erpApi.createVariant(body as unknown as Partial<ProductVariant>));
    },
    onSuccess: (_result, variables) =>
      qc.invalidateQueries({ queryKey: ['variants', variables.productId] }),
    onError: () => toast.error('Failed to create variant'),
  });
}

export interface VariantEditInput {
  id: number;
  productId: number;
  name: string;
  sku: string;
  price: number;
  stock: number;
  attributes: Record<string, string>;
}

export function useUpdateVariant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: VariantEditInput) =>
      unwrap(
        erpApi.updateVariant({
          id: data.id,
          productId: data.productId,
          name: data.name,
          sku: data.sku,
          salePrice: data.price,
          stockQuantity: data.stock,
          attributes: JSON.stringify(data.attributes ?? {}),
        })
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['variants'] }),
    onError: () => toast.error('Failed to update variant'),
  });
}

export function useDeleteVariant() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.deleteVariant(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['variants'] }),
    onError: () => toast.error('Failed to delete variant'),
  });
}
