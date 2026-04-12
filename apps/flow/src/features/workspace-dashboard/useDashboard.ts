import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

export function useFlowDashboard(workspaceId?: string) {
  return useQuery({
    queryKey: ['flow-dashboard', workspaceId],
    queryFn: () => unwrap(flowApi.getDashboard(workspaceId)),
    enabled: !!workspaceId,
  });
}
