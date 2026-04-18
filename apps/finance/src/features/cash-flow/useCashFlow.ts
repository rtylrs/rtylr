import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { cashFlowPointSchema } from '../../shared/schemas/cashFlow';
import type { CashFlowPoint } from '../../shared/components/CashFlowChart';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useCashFlow() {
  return useQuery({
    queryKey: ['cash-flow'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getCashFlow());
      return z
        .array(cashFlowPointSchema)
        .parse(unwrapList(res, 'points', 'cash_flow')) as CashFlowPoint[];
    },
  });
}
