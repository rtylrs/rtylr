import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi } from '../../shared/services/erpApi';
import { unwrapList } from '../../shared/lib/unwrapList';

/**
 * Backend price-list contract (snake_case -> camelCase via @voxire/api):
 * { id, ownerId, name, currency, isDefault, createdAt, updatedAt }.
 * There is NO tier / multiplier / activeCount / description on the backend, so
 * this is the real shape the UI works with. We expose this local row type
 * rather than the legacy shared `PriceList` interface (which still carries the
 * removed tier/multiplier fields).
 */
const priceListRowSchema = z.object({
  id: z.number(),
  name: z.string(),
  currency: z.string().optional().default(''),
  isDefault: z.boolean().optional().default(false),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export type PriceListRow = z.infer<typeof priceListRowSchema>;

export interface CreatePriceListInput {
  name: string;
  currency: string;
  isDefault: boolean;
}

export function usePriceLists() {
  return useQuery<PriceListRow[]>({
    queryKey: ['price-lists'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listPriceLists());
      return priceListRowSchema.array().parse(unwrapList(res, 'price_lists', 'priceLists'));
    },
  });
}

export function useCreatePriceList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreatePriceListInput) => unwrap(erpApi.createPriceList(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['price-lists'] }),
    onError: () => toast.error('Failed to create price list'),
  });
}
