import { oneKitResult, unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { insightsFinancialClient } from '../../shared/lib/clients';
import { cashFlowTrendPointSchema } from '../../shared/schemas/cashFlow';
import type { CashFlowPoint } from '../../shared/components/CashFlowChart';
import { unwrapList } from '../../shared/lib/unwrapList';

export interface UseCashFlowOptions {
  period?: 'month' | 'quarter' | 'year';
  /**
   * Ignored by the server, and kept only so CashFlow.tsx keeps compiling.
   *
   * GetCashFlowAnalyticsRequest accepts `period` and nothing else; the handler
   * derives the date range itself through resolveDateRangeFromPeriod. This hook
   * used to compute `from` and `to` from this date and send them, which the
   * server discarded. The date picker on the screen therefore changes nothing.
   *
   * Left in place rather than removed because deleting it means deciding what
   * that control should become, which is a UI change. Tracked in #910.
   */
  date?: Date;
}

export function useCashFlow(options?: UseCashFlowOptions) {
  const period = options?.period ?? 'month';

  return useQuery({
    queryKey: ['cash-flow', 'filter', { period }],
    queryFn: async (): Promise<CashFlowPoint[]> => {
      const res = await unwrap(
        oneKitResult(() => insightsFinancialClient.getCashFlowAnalytics({ period }))
      );

      // The response is { analytics: { inflows, outflows, net, trend } }, and
      // the trend points are already in major units: the handler applies
      // minorToMajor before returning (insights financial_service.go).
      const analytics = (res as { analytics?: unknown } | null)?.analytics ?? {};
      const trend = z.array(cashFlowTrendPointSchema).parse(unwrapList(analytics, 'trend'));

      return trend.map((p) => ({
        month: p.date,
        inflow: p.inflow,
        outflow: p.outflow,
      }));
    },
  });
}
