import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, PriceList } from '../../shared/services/erpApi';

export function usePriceLists() {
  return useQuery<PriceList[]>({
    queryKey: ['price-lists'],
    queryFn: async () => {
      const res = await erpApi.listPriceLists();
      return res.data ?? [];
    },
  });
}
