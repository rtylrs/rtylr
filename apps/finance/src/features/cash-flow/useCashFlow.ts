import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../shared/lib/api';
import { cashFlowSeriesPointSchema } from '../../shared/schemas/cashFlow';
import type { CashFlowPoint } from '../../shared/components/CashFlowChart';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useCashFlow() {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth() - 5, 1).toISOString().slice(0, 10);
  const to = now.toISOString().slice(0, 10);
  return useQuery({
    queryKey: ['cash-flow', 'filter', { from, to }],
    queryFn: async (): Promise<CashFlowPoint[]> => {
      // Backend cash-flow time series lives at /insights/financial/cash-flow and
      // returns { series: [{ date, inflowMinorUnits, outflowMinorUnits }] }.
      const res = await unwrap(
        safe.get<unknown>('/insights/financial/cash-flow', { params: { from, to } })
      );
      const series = z.array(cashFlowSeriesPointSchema).parse(unwrapList(res, 'series'));
      return series.map((p) => ({
        month: p.date,
        inflow: p.inflowMinorUnits / 100,
        outflow: p.outflowMinorUnits / 100,
      }));
    },
  });
}
