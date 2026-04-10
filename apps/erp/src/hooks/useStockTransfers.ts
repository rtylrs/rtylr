import { useQuery } from '@tanstack/react-query';
import { erpApi, StockTransfer } from '../services/erpApi';

export function useStockTransfers() {
  return useQuery<StockTransfer[]>({
    queryKey: ['stock-transfers'],
    queryFn: async () => {
      const res = await erpApi.listStockTransfers();
      return res.data ?? [];
    },
  });
}
