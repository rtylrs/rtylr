import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

export function useProjectOverview(projectId: string | undefined) {
  return useQuery({
    queryKey: ['project-overview', projectId],
    queryFn: async () => {
      if (!projectId) return null;
      const [taskList, ms] = await Promise.all([
        unwrap(flowApi.getTasks({ projectId })),
        unwrap(flowApi.getMilestones(projectId)),
      ]);
      const done = taskList.filter((t) => t.status === 'done').length;
      const inProgress = taskList.filter((t) => t.status === 'in_progress').length;
      const overdue = taskList.filter((t) => {
        if (!t.dueDate) return false;
        return new Date(t.dueDate) < new Date() && t.status !== 'done';
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
