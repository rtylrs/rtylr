import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { GetFinanceDashboardResponse } from '@voxire/voxie/finance/dashboard/v1/get_finance_dashboard';
import { safe } from '../../lib/api';

export const useFinanceDashboard = () => {
  return useQuery<GetFinanceDashboardResponse>({
    queryKey: ['finance-dashboard'],
    queryFn: async (): Promise<GetFinanceDashboardResponse> => {
      const res = await unwrap(safe.get<GetFinanceDashboardResponse>('/finance/dashboard'));
      return res as GetFinanceDashboardResponse;
    },
  });
};
