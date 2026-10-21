import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

export function useRoadmap() {
  return useQuery({
    queryKey: ['roadmap'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getRoadmap());
      return res.projects ?? [];
    },
  });
}
