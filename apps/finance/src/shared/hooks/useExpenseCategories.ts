import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financeApi, ExpenseCategory } from '../services/financeApi';
import { expenseCategorySchema } from '../schemas/expenses';

export function useExpenseCategories() {
  return useQuery({
    queryKey: ['expense-categories'],
    queryFn: async () => {
      const res = await unwrap(financeApi.getExpenseCategories());
      const raw = res ?? [];
      return z
        .array(expenseCategorySchema)
        .parse(Array.isArray(raw) ? raw : []) as ExpenseCategory[];
    },
  });
}
