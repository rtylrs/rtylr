import { useQuery } from '@tanstack/react-query';
import { MOCK_CATEGORIES, ExpenseCategoryBudget } from '../lib/mockData';

export function useExpenseCategories() {
  return useQuery<ExpenseCategoryBudget[]>({
    queryKey: ['expense-categories'],
    queryFn: async () => MOCK_CATEGORIES,
  });
}
