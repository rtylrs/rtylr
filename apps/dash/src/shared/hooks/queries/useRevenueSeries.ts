import { useQuery } from '@tanstack/react-query';
import api from '../../lib/api';

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
        const res = await api.get(`/insights/revenue`, {
          params: { from: isoDate(from), to: isoDate(today) },
        });
        const series = (res.data as { series?: RevenuePoint[] })?.series;
        return Array.isArray(series) ? series : [];
      } catch {
        // Endpoint may not be deployed yet (insights namespace is P3
        // gap-fill in progress). Fail open — the widget shows the
        // "no data" state, not an error.
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
