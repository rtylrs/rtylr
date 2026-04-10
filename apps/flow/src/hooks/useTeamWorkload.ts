import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useTeamWorkload() {
  return useQuery({
    queryKey: ['team-workload'],
    queryFn: () => flowApi.getTeamWorkload(),
  });
}
