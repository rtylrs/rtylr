import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiTaskActivity } from '../../shared/services/flowApi';

export function useTaskActivity(taskId: string) {
  return useQuery<ApiTaskActivity[]>({
    queryKey: ['task-activity', taskId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getTaskActivity(taskId));
      return res.activity;
    },
    enabled: !!taskId,
  });
}
