import { useQuery } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { RevenueServiceClient } from '@voxire/voxie/insights/revenue/v1/client';

/**
 * Revenue time-series for the dashboard widget. Hits
 * insights.revenue.v1.RevenueService.GetRevenueOverview, which returns one
 * row per day in the window. Empty data → empty array, so the widget can
 * render the same "no activity" state whether the endpoint is unavailable
 * or the tenant just hasn't made sales yet.
 */

const revenueClient = createOneKitClient(RevenueServiceClient, 'dash');

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
        const res = await revenueClient.getRevenueOverview({
          period: `${isoDate(from)},${isoDate(today)}`,
        });
        const trend = res.overview?.trend ?? [];
        return trend.map((point) => ({
          date: point.date ?? '',
          amountMinorUnits: Math.round((point.amount ?? 0) * 100),
        }));
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
