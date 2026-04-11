import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { erpApi, StockTransfer } from '../../shared/services/erpApi';

export function useStockTransfers() {
  return useQuery<StockTransfer[]>({
    queryKey: ['stock-transfers'],
    queryFn: () => unwrap(erpApi.listStockTransfers()),
  });
}
