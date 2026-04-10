import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useRoadmap() {
  return useQuery({
    queryKey: ['roadmap'],
    queryFn: () => flowApi.getRoadmap(),
  });
}
