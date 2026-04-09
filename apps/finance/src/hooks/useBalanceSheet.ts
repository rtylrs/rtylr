import { useQuery } from '@tanstack/react-query';
import { MOCK_BALANCE_SHEET, BalanceSheet } from '../lib/mockData';

export function useBalanceSheet() {
  return useQuery<BalanceSheet>({
    queryKey: ['balance-sheet'],
    queryFn: async () => MOCK_BALANCE_SHEET,
  });
}
