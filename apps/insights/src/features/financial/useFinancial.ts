import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';

export function useProfitLoss(params?: { start?: string; end?: string }) {
  return useQuery({
    queryKey: ['insights', 'pl', params],
    queryFn: () => unwrap(insightsService.getProfitLoss(params)),
  });
}

export function useExpenseBreakdown(params?: { start?: string; end?: string }) {
  return useQuery({
    queryKey: ['insights', 'expenses', params],
    queryFn: () => unwrap(insightsService.getExpenseBreakdown(params)),
  });
}

export function useInvoiceAging() {
  return useQuery({
    queryKey: ['insights', 'invoice-aging'],
    queryFn: () => unwrap(insightsService.getInvoiceAging()),
  });
}

export function useCashFlow(period?: string) {
  return useQuery({
    queryKey: ['insights', 'cash-flow', period],
    queryFn: () => unwrap(insightsService.getCashFlow({ period })),
  });
}
