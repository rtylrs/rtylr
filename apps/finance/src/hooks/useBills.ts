import { useQuery } from '@tanstack/react-query';
import { MOCK_BILLS, Bill } from '../lib/mockData';

export function useBills() {
  return useQuery<Bill[]>({
    queryKey: ['bills'],
    queryFn: async () => MOCK_BILLS,
  });
}
