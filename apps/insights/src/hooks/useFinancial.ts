import { useQuery } from '@tanstack/react-query';
import { insightsService } from '../services/insightsService';

export function useProfitLoss(params?: { start?: string; end?: string }) {
  return useQuery({
    queryKey: ['insights', 'pl', params],
    queryFn: () => insightsService.getProfitLoss(params).then((r) => r.data),
  });
}

export function useExpenseBreakdown(params?: { start?: string; end?: string }) {
  return useQuery({
    queryKey: ['insights', 'expenses', params],
    queryFn: () => insightsService.getExpenseBreakdown(params).then((r) => r.data),
  });
}

export function useInvoiceAging() {
  return useQuery({
    queryKey: ['insights', 'invoice-aging'],
    queryFn: () => insightsService.getInvoiceAging().then((r) => r.data),
  });
}

export function useCashFlow(period?: string) {
  return useQuery({
    queryKey: ['insights', 'cash-flow', period],
    queryFn: () => insightsService.getCashFlow({ period }).then((r) => r.data),
  });
}
