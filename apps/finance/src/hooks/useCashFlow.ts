import { useQuery } from '@tanstack/react-query';
import { MOCK_CASH_FLOW, CashFlowMonth } from '../lib/mockData';

export function useCashFlow() {
  return useQuery<CashFlowMonth[]>({
    queryKey: ['cash-flow'],
    queryFn: async () => MOCK_CASH_FLOW,
  });
}
