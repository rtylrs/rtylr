import { useQuery } from '@tanstack/react-query';
import { dashboardClient } from '../../lib/clients';
import { dashKpisSchema } from '../../schemas/kpis';

export type { HubKpis } from '../../schemas/kpis';

export const useKpis = (range: string = '7d') => {
  return useQuery({
    queryKey: ['dash-kpis', range],
    queryFn: async () => {
      const res = await dashboardClient.getKPIs({ range });
      return dashKpisSchema.parse(res);
    },
    staleTime: 30_000,
  });
};
