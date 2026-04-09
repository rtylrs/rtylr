import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useFlowDashboard(workspaceId?: string) {
  return useQuery({
    queryKey: ['flow-dashboard', workspaceId],
    queryFn: () => flowApi.getDashboard(workspaceId),
  });
}
