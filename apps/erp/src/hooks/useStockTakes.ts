import { useQuery } from '@tanstack/react-query';
import { erpApi, StockTake } from '../services/erpApi';

export function useStockTakes() {
  return useQuery<StockTake[]>({
    queryKey: ['stock-takes'],
    queryFn: async () => {
      const res = await erpApi.listStockTakes();
      return res.data ?? [];
    },
  });
}
