import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, Brand } from '../../shared/services/erpApi';
import { unwrapList } from '../../shared/lib/unwrapList';

/**
 * Backend brand contract (snake_case -> camelCase via @voxire/api):
 * { id, ownerId, name, logoUrl, description, productCount, createdAt, updatedAt }.
 * createdAt / updatedAt are returned but not displayed yet; productCount is
 * genuinely returned by the backend, so it stays.
 */
const brandRowSchema = z.object({
  id: z.number(),
  name: z.string(),
  logoUrl: z.string().optional().nullable(),
  description: z.string().optional().nullable(),
  productCount: z.number().default(0),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export function useBrands() {
  return useQuery<Brand[]>({
    queryKey: ['brands'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listBrands());
      const rows = z.array(brandRowSchema).parse(unwrapList(res, 'brands'));
      return rows.map(
        (row): Brand => ({
          id: row.id,
          name: row.name,
          logoUrl: row.logoUrl ?? undefined,
          description: row.description ?? undefined,
          productCount: row.productCount,
        })
      );
    },
  });
}

export function useCreateBrand() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Brand>) => unwrap(erpApi.createBrand(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['brands'] }),
    onError: () => toast.error('Failed to create brand'),
  });
}

export function useDeleteBrand() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(erpApi.deleteBrand(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['brands'] }),
    onError: () => toast.error('Failed to delete brand'),
  });
}
