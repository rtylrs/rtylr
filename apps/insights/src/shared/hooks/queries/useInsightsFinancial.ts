import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type { CashFlowResponse, ExpensesByCategoryResponse, InvoiceAgingResponse } from './types';

export const useCashFlow = (params?: DateRangeParams): UseQueryResult<CashFlowResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<CashFlowResponse>({
    queryKey: ['insights-financial-cash-flow', { from, to }],
    queryFn: async (): Promise<CashFlowResponse> => {
      const res = await unwrap(
        safe.get<CashFlowResponse>('/insights/financial/cash-flow', { params: { from, to } })
      );
      return res as CashFlowResponse;
    },
  });
};

export const useExpensesByCategory = (
  params?: DateRangeParams
): UseQueryResult<ExpensesByCategoryResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<ExpensesByCategoryResponse>({
    queryKey: ['insights-financial-expenses', { from, to }],
    queryFn: async (): Promise<ExpensesByCategoryResponse> => {
      const res = await unwrap(
        safe.get<ExpensesByCategoryResponse>('/insights/financial/expenses', {
          params: { from, to },
        })
      );
      return res as ExpensesByCategoryResponse;
    },
  });
};

export const useInvoiceAging = (): UseQueryResult<InvoiceAgingResponse> => {
  return useQuery<InvoiceAgingResponse>({
    queryKey: ['insights-financial-invoice-aging'],
    queryFn: async (): Promise<InvoiceAgingResponse> => {
      const res = await unwrap(safe.get<InvoiceAgingResponse>('/insights/financial/invoice-aging'));
      return res as InvoiceAgingResponse;
    },
  });
};
