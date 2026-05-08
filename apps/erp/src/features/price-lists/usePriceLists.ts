import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { erpApi, PriceList } from '../../shared/services/erpApi';
import { priceListSchema } from '../../shared/schemas/priceLists';
import { unwrapList } from '../../shared/lib/unwrapList';

export function usePriceLists() {
  return useQuery<PriceList[]>({
    queryKey: ['price-lists'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listPriceLists());
      return z
        .array(priceListSchema)
        .parse(unwrapList(res, 'price_lists', 'priceLists')) as PriceList[];
    },
  });
}

export function useCreatePriceList() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<PriceList>) => unwrap(erpApi.createPriceList(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['price-lists'] }),
    onError: () => toast.error('Failed to create price list'),
  });
}
