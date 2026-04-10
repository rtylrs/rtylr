import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useInsights(workspaceId?: string) {
  return useQuery({
    queryKey: ['insights', workspaceId],
    queryFn: () => flowApi.getInsights(workspaceId),
  });
}
