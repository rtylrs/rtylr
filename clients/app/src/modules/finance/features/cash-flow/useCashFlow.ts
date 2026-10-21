import { oneKitResult, unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { insightsFinancialClient } from '../../shared/lib/clients';
import { cashFlowTrendPointSchema } from '../../shared/schemas/cashFlow';
import type { CashFlowPoint } from '../../shared/components/CashFlowChart';
import { unwrapList } from '../../shared/lib/unwrapList';

export interface UseCashFlowOptions {
  /**
   * Reporting granularity. GetCashFlowAnalyticsRequest accepts `period` and
   * nothing else: the server derives the date range itself through
   * resolveDateRangeFromPeriod, so there is deliberately no date option here.
   * A user-picked range needs a contract change first (#910).
   */
  period?: 'month' | 'quarter' | 'year';
}

export function useCashFlow(options?: UseCashFlowOptions) {
  const period = options?.period ?? 'month';

  return useQuery({
    queryKey: ['finance', 'cash-flow', 'filter', { period }],
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
