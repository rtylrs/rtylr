import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

export function useTeamWorkload() {
  return useQuery({
    queryKey: ['team-workload'],
    queryFn: () => unwrap(flowApi.getTeamWorkload()),
  });
}
