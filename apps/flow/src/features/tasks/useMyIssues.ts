import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';
import type { ApiTask } from '../../shared/services/flowApi';

/** Tasks assigned to the current user (My tasks). */
export function useMyIssues() {
  return useQuery<ApiTask[]>({
    queryKey: ['my-tasks'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getMyTasks());
      return unwrapListResponse<ApiTask>(res, 'tasks');
    },
  });
}
