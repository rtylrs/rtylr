import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi } from '../../shared/services/crmApi';
import { dashboardStatsSchema } from '../../shared/schemas/dashboard';

export function useCRMDashboard() {
  return useQuery({
    queryKey: ['crm-dashboard'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getDashboard());
      return dashboardStatsSchema.parse(res);
    },
  });
}
