import { useQuery } from '@tanstack/react-query';
import { dashboardClient } from '../../lib/clients';
import { dashKpisSchema } from '../../schemas/kpis';

export type { HubKpis } from '../../schemas/kpis';

/**
 * No range parameter. hub.dashboard.v1.GetKPIsRequest is an empty message, so
 * `{ range }` compiled only because excess-property checks do not apply to one,
 * and the generated encoder emitted `{}`. The range reached the queryKey and
 * nothing else, so switching ranges refetched byte-identical data under a new
 * key (rtylr#946).
 *
 * The figures are today-scoped server-side: onion's kpis.sql defines revenue and
 * orderCount as SumRevenueToday and CountOrdersToday.
 */
export const useKpis = () => {
  return useQuery({
    queryKey: ['dash-kpis'],
    queryFn: async () => {
      const res = await dashboardClient.getKPIs({});
      return dashKpisSchema.parse(res);
    },
    staleTime: 30_000,
  });
};
