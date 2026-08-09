import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

export function useInsights(workspaceId?: string) {
  return useQuery({
    queryKey: ['insights', workspaceId],
    queryFn: () => unwrap(flowApi.getInsights(workspaceId)),
    enabled: !!workspaceId,
  });
}
