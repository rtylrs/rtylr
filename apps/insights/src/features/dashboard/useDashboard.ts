import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';

export function useKpis() {
  return useQuery({
    queryKey: ['insights', 'kpis'],
    queryFn: () => unwrap(insightsService.getKpis()),
  });
}

export function useDashboardSummary() {
  return useQuery({
    queryKey: ['insights', 'dashboard'],
    queryFn: () => unwrap(insightsService.getDashboardSummary()),
  });
}
