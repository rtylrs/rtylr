import { useQuery } from '@tanstack/react-query';
import type { GetFinanceDashboardResponse } from '@voxire/voxie/finance/dashboard/v1/types';
import { financeDashboardClient } from '../../lib/clients';

export const useFinanceDashboard = () => {
  return useQuery<GetFinanceDashboardResponse>({
    queryKey: ['finance-dashboard'],
    queryFn: async (): Promise<GetFinanceDashboardResponse> => {
      return financeDashboardClient.getFinanceDashboard({});
    },
  });
};
