import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, StockTransfer } from '../../shared/services/erpApi';
import { stockTransferSchema } from '../../shared/schemas/stockTransfers';

export function useStockTransfers() {
  return useQuery<StockTransfer[]>({
    queryKey: ['stock-transfers'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listStockTransfers());
      return z.array(stockTransferSchema).parse(Array.isArray(res) ? res : []) as StockTransfer[];
    },
  });
}
