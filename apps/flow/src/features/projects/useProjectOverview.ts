import { unwrap } from '@voxire/api';
import { isOverdue } from '../../shared/lib/dateUtils';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiTask, ApiMilestone } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useProjectOverview(projectId: string | undefined) {
  return useQuery({
    queryKey: ['project-overview', projectId],
    queryFn: async () => {
      if (!projectId) return null;
      const [tasksRes, msRes] = await Promise.all([
        unwrap(flowApi.getTasks({ projectId })),
        unwrap(flowApi.getMilestones(projectId)),
      ]);
      const taskList = unwrapListResponse<ApiTask>(tasksRes, 'tasks');
      const ms = unwrapListResponse<ApiMilestone>(msRes, 'milestones');
      const done = taskList.filter((t) => t.status === 'done').length;
      const inProgress = taskList.filter((t) => t.status === 'in_progress').length;
      const overdue = taskList.filter((t) => {
        if (!t.dueDate) return false;
        return isOverdue(t.dueDate, t.status);
      }).length;
      return {
        tasks: taskList,
        stats: {
          total: taskList.length,
          done,
          inProgress,
          overdue,
          progress: taskList.length ? Math.round((done / taskList.length) * 100) : 0,
        },
        milestones: ms,
      };
    },
    enabled: !!projectId,
  });
}
