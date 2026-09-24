import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import type { Expense as ApiExpense } from '@voxire/voxie/finance/expenses/v1/types';
import type { Expense } from '../../shared/services/financeApi';

const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

export interface ExpenseDetail extends Expense {
  paymentMethod?: string | null;
  isTaxDeductible?: boolean | null;
  taxAmount?: number | null;
}

function toNullableNumber(value: unknown): number | null {
  if (value === null || value === undefined || value === '') return null;
  return toFiniteNumber(value);
}

export function mapExpense(expense: ApiExpense | undefined): ExpenseDetail | undefined {
  if (!expense) return undefined;

  return {
    id: expense.id ?? '',
    ownerId: toNullableNumber(expense.ownerId),
    categoryId: toNullableNumber(expense.categoryId),
    description: expense.description ?? '',
    amount: toFiniteNumber(expense.amount),
    currency: expense.currency,
    vendor: expense.vendor,
    receiptUrl: expense.receiptUrl,
    paymentMethod: expense.paymentMethod ?? null,
    isTaxDeductible: expense.isTaxDeductible ?? null,
    taxAmount: toNullableNumber(expense.taxAmount),
    notes: expense.notes ?? null,
    expenseDate: expense.expenseDate ?? '',
    status: expense.status ?? '',
    createdAt: expense.createdAt,
    updatedAt: expense.updatedAt,
  };
}

export function useExpenseById(id: string | undefined) {
  return useQuery<ExpenseDetail>({
    queryKey: ['expense', id],
    queryFn: async (): Promise<ExpenseDetail> => {
      const res = await expensesClient.getExpense({ id: id ?? '' });
      const expense = mapExpense(res.expense);
      if (!expense) throw new Error('Expense not found');
      return expense;
    },
    enabled: !!id,
  });
}
