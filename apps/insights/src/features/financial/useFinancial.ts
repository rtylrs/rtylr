import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';
import {
  profitLossSchema,
  expenseBreakdownSchema,
  invoiceAgingSchema,
  cashFlowSchema,
} from '../../shared/schemas/financial';

export function useProfitLoss(params?: { start?: string; end?: string }) {
  return useQuery({
    queryKey: ['insights', 'pl', params],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getProfitLoss(params));
      return profitLossSchema.parse(res);
    },
  });
}

export function useExpenseBreakdown(params?: { start?: string; end?: string }) {
  return useQuery({
    queryKey: ['insights', 'expenses', params],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getExpenseBreakdown(params));
      return expenseBreakdownSchema.parse(res);
    },
  });
}

export function useInvoiceAging() {
  return useQuery({
    queryKey: ['insights', 'invoice-aging'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getInvoiceAging());
      return invoiceAgingSchema.parse(res);
    },
  });
}

export function useCashFlow(period?: string) {
  return useQuery({
    queryKey: ['insights', 'cash-flow', period],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getCashFlow({ period }));
      return cashFlowSchema.parse(res);
    },
  });
}
