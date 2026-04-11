import { useQuery } from '@tanstack/react-query';
import { insightsService } from '../services/insightsService';

export function useKpis() {
  return useQuery({
    queryKey: ['insights', 'kpis'],
    queryFn: () => insightsService.getKpis().then((r) => r.data),
  });
}

export function useDashboardSummary() {
  return useQuery({
    queryKey: ['insights', 'dashboard'],
    queryFn: () => insightsService.getDashboardSummary().then((r) => r.data),
  });
}
