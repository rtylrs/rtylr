import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { erpApi, StockTake } from '../../shared/services/erpApi';
import { stockTakeSchema } from '../../shared/schemas/stockTakes';

export function useStockTakes() {
  return useQuery<StockTake[]>({
    queryKey: ['stock-takes'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listStockTakes());
      return z.array(stockTakeSchema).parse(Array.isArray(res) ? res : []) as StockTake[];
    },
  });
}
