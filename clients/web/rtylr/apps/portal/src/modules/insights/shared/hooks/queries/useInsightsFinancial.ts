import { createOneKitClient } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { FinancialServiceClient } from '@voxire/voxie/insights/financial/v1/client';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type { CashFlowResponse, ExpensesByCategoryResponse, InvoiceAgingResponse } from './types';

const financialClient = createOneKitClient(FinancialServiceClient, 'insights');

const majorToMinor = (value: number): number => Math.round(value * 100);

export const useCashFlow = (params?: DateRangeParams): UseQueryResult<CashFlowResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<CashFlowResponse>({
    queryKey: ['insights-financial-cash-flow', { from, to }],
    queryFn: async (): Promise<CashFlowResponse> => {
      const res = await financialClient.getCashFlowAnalytics({ period: `${from},${to}` });
      const trend = res.analytics?.trend ?? [];
      return {
        series: trend.map((point) => ({
          date: point.date ?? '',
          inflowMinorUnits: majorToMinor(point.inflow ?? 0),
          outflowMinorUnits: majorToMinor(point.outflow ?? 0),
        })),
      };
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
      const res = await financialClient.getExpenseAnalytics({ period: `${from},${to}` });
      const byCategory = res.analytics?.byCategory ?? [];
      return {
        categories: byCategory.map((entry, index) => ({
          categoryId: index + 1,
          categoryName: entry.category ?? '',
          amountMinorUnits: majorToMinor(entry.amount ?? 0),
        })),
      };
    },
  });
};

export const useInvoiceAging = (): UseQueryResult<InvoiceAgingResponse> => {
  return useQuery<InvoiceAgingResponse>({
    queryKey: ['insights-financial-invoice-aging'],
    queryFn: async (): Promise<InvoiceAgingResponse> => {
      const res = await financialClient.getInvoiceAging({});
      const buckets = res.aging?.buckets ?? [];
      return {
        buckets: buckets.map((b) => ({
          bucket: b.bucket ?? '',
          invoiceCount: b.count ?? 0,
          amountMinorUnits: majorToMinor(b.amount ?? 0),
        })),
      };
    },
  });
};
