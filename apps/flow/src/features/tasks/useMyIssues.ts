import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

/** Tasks assigned to the current user (My tasks). */
export function useMyIssues() {
  return useQuery({
    queryKey: ['my-tasks'],
    queryFn: () => unwrap(flowApi.getMyTasks()),
  });
}
