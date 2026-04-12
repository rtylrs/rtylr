import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';
import { insightsKpisSchema, insightsDashboardSummarySchema } from '../../shared/schemas/dashboard';

export function useKpis() {
  return useQuery({
    queryKey: ['insights', 'kpis'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getKpis());
      return insightsKpisSchema.parse(res);
    },
  });
}

export function useDashboardSummary() {
  return useQuery({
    queryKey: ['insights', 'dashboard'],
    queryFn: async (): Promise<any> => {
      const res = await unwrap(insightsService.getDashboardSummary());
      return insightsDashboardSummarySchema.parse(res);
    },
  });
}
