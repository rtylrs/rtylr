import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, StockTake } from '../../shared/services/erpApi';

export function useStockTakes() {
  return useQuery<StockTake[]>({
    queryKey: ['stock-takes'],
    queryFn: () => unwrap(erpApi.listStockTakes()),
  });
}
