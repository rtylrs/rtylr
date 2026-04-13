import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';

export function useTeamWorkload() {
  const { activeWorkspaceId } = useFlowStore();
  return useQuery({
    queryKey: ['team-workload', activeWorkspaceId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getTeamWorkload(activeWorkspaceId));
      return res.members ?? [];
    },
  });
}
