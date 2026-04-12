import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../../shared/services/financeApi';
import { cashFlowPointSchema } from '../../shared/schemas/cashFlow';
import type { CashFlowPoint } from '../../shared/components/CashFlowChart';

export function useCashFlow() {
  return useQuery({
    queryKey: ['cash-flow'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getCashFlow());
      const raw = res ?? [];
      return z.array(cashFlowPointSchema).parse(Array.isArray(raw) ? raw : []) as CashFlowPoint[];
    },
  });
}
