import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import { cashFlowPointSchema } from '../../shared/schemas/cashFlow';
import type { CashFlowPoint } from '../../shared/components/CashFlowChart';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useCashFlow() {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth() - 5, 1).toISOString().slice(0, 10);
  const to = now.toISOString().slice(0, 10);
  return useQuery({
    queryKey: ['cash-flow', { from, to }],
    queryFn: async (): Promise<CashFlowPoint[]> => {
      const res = await unwrap(
        safe.get<unknown>('/finance/reports/cash-flow', { params: { from, to } })
      );
      return z
        .array(cashFlowPointSchema)
        .parse(unwrapList(res, 'points', 'cash_flow')) as CashFlowPoint[];
    },
  });
}
