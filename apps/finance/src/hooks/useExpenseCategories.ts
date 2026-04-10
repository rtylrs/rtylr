import { useQuery } from '@tanstack/react-query';
import { financeApi } from '../services/financeApi';

export function useExpenseCategories() {
  return useQuery({
    queryKey: ['expense-categories'],
    queryFn: () => financeApi.getExpenseCategories(),
  });
}
