import { useQuery } from '@tanstack/react-query';
import { erpApi, StockTake } from '../services/erpApi';
import { mockStockTakes } from '../lib/mockData';

export function useStockTakes() {
  return useQuery<StockTake[]>({
    queryKey: ['stock-takes'],
    queryFn: async () => {
      try {
        const res = await erpApi.listStockTakes();
        if (res.data && res.data.length) return res.data;
        return mockStockTakes;
      } catch {
        return mockStockTakes;
      }
    },
  });
}
