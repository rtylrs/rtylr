import { oneKitResult, unwrap } from '@voxire/api';
import { z } from 'zod';
import { useQuery } from '@tanstack/react-query';
import { financialCompatibilityClient } from '../../shared/lib/clients';
import { cashFlowSeriesPointSchema } from '../../shared/schemas/cashFlow';
import type { CashFlowPoint } from '../../shared/components/CashFlowChart';
import { unwrapList } from '../../shared/lib/unwrapList';

import { minorToMajor } from '../../shared/lib/money';

export interface UseCashFlowOptions {
  period?: 'month' | 'quarter' | 'year';
  date?: Date;
}

export function useCashFlow(options?: UseCashFlowOptions) {
  const period = options?.period ?? 'month';
  const baseDate = options?.date ?? new Date();

  let from: string;
  let to: string;

  if (period === 'month') {
    const start = new Date(baseDate.getFullYear(), baseDate.getMonth(), 1);
    const end = new Date(baseDate.getFullYear(), baseDate.getMonth() + 1, 0);
    from = start.toISOString().slice(0, 10);
    to = end.toISOString().slice(0, 10);
  } else if (period === 'quarter') {
    const q = Math.floor(baseDate.getMonth() / 3);
    const start = new Date(baseDate.getFullYear(), q * 3, 1);
    const end = new Date(baseDate.getFullYear(), (q + 1) * 3, 0);
    from = start.toISOString().slice(0, 10);
    to = end.toISOString().slice(0, 10);
  } else {
    const start = new Date(baseDate.getFullYear(), 0, 1);
    const end = new Date(baseDate.getFullYear(), 11, 31);
    from = start.toISOString().slice(0, 10);
    to = end.toISOString().slice(0, 10);
  }

  return useQuery({
    queryKey: ['cash-flow', 'filter', { from, to, period }],
    queryFn: async (): Promise<CashFlowPoint[]> => {
      // Backend cash-flow time series lives at /insights/financial/cash-flow and
      // returns { series: [{ date, inflowMinorUnits, outflowMinorUnits }] }.
      const res = await unwrap(
        oneKitResult(() => financialCompatibilityClient.getCashFlow({ from, to }))
      );
      const series = z.array(cashFlowSeriesPointSchema).parse(unwrapList(res, 'series'));
      return series.map((p) => ({
        month: p.date,
        inflow: minorToMajor(Math.round(p.inflowMinorUnits)),
        outflow: minorToMajor(Math.round(p.outflowMinorUnits)),
      }));
    },
  });
}
