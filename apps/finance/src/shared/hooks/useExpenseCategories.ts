import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../lib/api';
import type { ExpenseCategory } from '../services/financeApi';
import { expenseCategorySchema } from '../schemas/expenses';
import { unwrapList } from '../lib/unwrapList';

export function useExpenseCategories() {
  return useQuery({
    queryKey: ['expense-categories'],
    queryFn: async (): Promise<ExpenseCategory[]> => {
      const res = await unwrap(safe.get<unknown>('/finance/expenses/categories'));
      return z
        .array(expenseCategorySchema)
        .parse(unwrapList(res, 'categories')) as ExpenseCategory[];
    },
  });
}
