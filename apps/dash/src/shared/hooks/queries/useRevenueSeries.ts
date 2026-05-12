import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { safe } from '../../lib/api';

/**
 * Revenue time-series for the dashboard widget. Hits the insights
 * aggregation endpoint (`GET /insights/revenue?from=&to=`), which
 * returns one row per day in the window. Empty data → empty array,
 * so the widget can render the same "no activity" state whether the
 * endpoint is unavailable or the tenant just hasn't made sales yet.
 *
 * API shape: { series: [{ date: 'YYYY-MM-DD', amount_minor_units: number }] }
 * The api layer converts snake_case to camelCase, so we consume
 * `amountMinorUnits` here.
 */

export interface RevenuePoint {
  date: string;
  amountMinorUnits: number;
}

function isoDate(d: Date): string {
  return d.toISOString().slice(0, 10);
}

export function useRevenueSeries(days = 7): {
  data: RevenuePoint[];
  isLoading: boolean;
} {
  const today = new Date();
  const from = new Date(today);
  from.setDate(from.getDate() - (days - 1));

  const query = useQuery({
    queryKey: ['insights', 'revenue', { days }],
    queryFn: async (): Promise<RevenuePoint[]> => {
      try {
        const res = await unwrap(
          safe.get<{ series?: RevenuePoint[] }>('/insights/revenue', {
            params: { from: isoDate(from), to: isoDate(today) },
          })
        );
        return Array.isArray(res.series) ? res.series : [];
      } catch {
        return [];
      }
    },
    staleTime: 60_000,
  });

  return {
    data: query.data ?? [],
    isLoading: query.isLoading,
  };
}
