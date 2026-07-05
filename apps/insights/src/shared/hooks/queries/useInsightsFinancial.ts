import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type { CashFlowResponse, ExpensesByCategoryResponse, InvoiceAgingResponse } from './types';
import { mapCashFlowAnalytics, mapExpenseAnalytics, mapInvoiceAging } from '../../utils/onekitMappers';

export const useCashFlow = (params?: DateRangeParams): UseQueryResult<CashFlowResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<CashFlowResponse>({
    queryKey: ['insights-financial-cash-flow', { from, to }],
    queryFn: async (): Promise<CashFlowResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/financial/v1/cash-flow', {
          params: { period: `${from},${to}` },
        })
      );
      return mapCashFlowAnalytics(res as Record<string, unknown>);
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
        safe.get<Record<string, unknown>>('/api/insights/financial/v1/expenses', {
          params: { period: `${from},${to}` },
        })
      );
      return mapExpenseAnalytics(res as Record<string, unknown>);
    },
  });
};

export const useInvoiceAging = (): UseQueryResult<InvoiceAgingResponse> => {
  return useQuery<InvoiceAgingResponse>({
    queryKey: ['insights-financial-invoice-aging'],
    queryFn: async (): Promise<InvoiceAgingResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/financial/v1/invoice-aging')
      );
      return mapInvoiceAging(res as Record<string, unknown>);
    },
  });
};
