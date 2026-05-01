import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { erpApi, StockTransfer } from '../../shared/services/erpApi';
import { stockTransferSchema } from '../../shared/schemas/stockTransfers';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useStockTransfers() {
  return useQuery<StockTransfer[]>({
    queryKey: ['stock-transfers'],
    queryFn: async () => {
      const res = await unwrap(erpApi.listStockTransfers());
      return z
        .array(stockTransferSchema)
        .parse(unwrapList(res, 'stock_transfers', 'stockTransfers')) as StockTransfer[];
    },
  });
}

export function useCreateStockTransfer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<StockTransfer>) => unwrap(erpApi.createStockTransfer(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['stock-transfers'] }),
  });
}
