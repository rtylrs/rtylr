import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
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
