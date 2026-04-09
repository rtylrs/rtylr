import { useQuery } from '@tanstack/react-query';
import { erpApi, PriceList } from '../services/erpApi';
import { mockPriceLists } from '../lib/mockData';

export function usePriceLists() {
  return useQuery<PriceList[]>({
    queryKey: ['price-lists'],
    queryFn: async () => {
      try {
        const res = await erpApi.listPriceLists();
        if (res.data && res.data.length) return res.data;
        return mockPriceLists;
      } catch {
        return mockPriceLists;
      }
    },
  });
}
