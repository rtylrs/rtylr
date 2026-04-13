import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';

export function useRoadmap() {
  const { activeWorkspaceId } = useFlowStore();
  return useQuery({
    queryKey: ['roadmap', activeWorkspaceId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getRoadmap(activeWorkspaceId));
      return res.projects ?? [];
    },
  });
}
