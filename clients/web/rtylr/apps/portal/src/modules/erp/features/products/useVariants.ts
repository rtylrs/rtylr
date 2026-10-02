import { createOneKitClient } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { VariantServiceClient } from '@voxire/voxie/erp/variant/v1/client';
import type {
  CreateVariantRequest,
  CreateVariantResponse,
  DeleteVariantResponse,
  UpdateVariantRequest,
  UpdateVariantResponse,
  Variant,
} from '@voxire/voxie/erp/variant/v1/types';
import { ERP_LIST_LIMIT } from '../../shared/lib/erpListBounds';

const variantClient = createOneKitClient(VariantServiceClient, 'erp');

export type { CreateVariantRequest, UpdateVariantRequest, Variant };

export function parseVariantAttributes(raw: string): Record<string, string> {
  if (!raw) return {};
  try {
    const parsed: unknown = JSON.parse(raw);
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {};
    return Object.fromEntries(
      Object.entries(parsed as Record<string, unknown>).map(([key, value]) => [key, String(value)])
    );
  } catch {
    return {};
  }
}

export async function fetchAllVariants(productId?: string): Promise<Variant[]> {
  const variants: Variant[] = [];
  const seenPages = new Set<string>();
  let offset = 0;

  while (true) {
    const response = await variantClient.listVariants({
      productId: productId ?? '0',
      limit: ERP_LIST_LIMIT,
      offset,
    });
    const page = response.variants ?? [];
    const signature = page
      .map((variant) => variant.id ?? `${variant.name}:${variant.sku}`)
      .join('|');
    if (page.length > 0 && seenPages.has(signature)) {
      throw new Error('Variant pagination did not advance');
    }
    if (page.length > 0) seenPages.add(signature);
    variants.push(...page);

    const hasMore =
      response.hasMore ??
      (response.total !== undefined
        ? variants.length < response.total
        : page.length === ERP_LIST_LIMIT);
    if (!hasMore || page.length === 0) return variants;
    offset += page.length;
  }
}

export function useVariants(productId?: string) {
  return useQuery<Variant[]>({
    queryKey: ['variants', productId],
    queryFn: () => fetchAllVariants(productId),
  });
}

export function useCreateVariant() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<CreateVariantResponse, Error, CreateVariantRequest>({
    mutationFn: (data) => variantClient.createVariant(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['variants'] }),
    onError: () => toast.error(t('erp.products.variants.createFailedOne')),
  });
}

export function useUpdateVariant() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<UpdateVariantResponse, Error, UpdateVariantRequest>({
    mutationFn: (data) => variantClient.updateVariant(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['variants'] }),
    onError: () => toast.error(t('erp.products.variants.updateFailed')),
  });
}

export function useDeleteVariant() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation<DeleteVariantResponse, Error, string>({
    mutationFn: (id) => variantClient.deleteVariant({ id }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['variants'] }),
    onError: () => toast.error(t('erp.products.variants.deleteFailed')),
  });
}
