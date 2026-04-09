import { useQuery } from '@tanstack/react-query';
import { erpApi, StockTransfer } from '../services/erpApi';
import { mockStockTransfers } from '../lib/mockData';

export function useStockTransfers() {
  return useQuery<StockTransfer[]>({
    queryKey: ['stock-transfers'],
    queryFn: async () => {
      try {
        const res = await erpApi.listStockTransfers();
        if (res.data && res.data.length) return res.data;
        return mockStockTransfers;
      } catch {
        return mockStockTransfers;
      }
    },
  });
}
