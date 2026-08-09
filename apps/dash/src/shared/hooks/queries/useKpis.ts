import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { dashKpisSchema } from '../../schemas/kpis';

export type { HubKpis } from '../../schemas/kpis';

export const useKpis = () => {
  return useQuery({
    queryKey: ['dash-kpis'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/hub/dashboard/v1/kpis'));
      return dashKpisSchema.parse(res);
    },
    staleTime: 30_000,
  });
};
