import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useProjectOverview(projectId: string | undefined) {
  return useQuery({
    queryKey: ['project-overview', projectId],
    queryFn: async () => {
      if (!projectId) return null;
      const [tasks, milestones] = await Promise.all([
        flowApi.getTasks({ projectId }),
        flowApi.getMilestones(projectId),
      ]);
      const done = tasks.filter((t) => t.status === 'done').length;
      const inProgress = tasks.filter((t) => t.status === 'in_progress').length;
      const overdue = tasks.filter((t) => {
        if (!t.dueDate) return false;
        return new Date(t.dueDate) < new Date() && t.status !== 'done';
      }).length;
      return {
        tasks,
        stats: {
          total: tasks.length,
          done,
          inProgress,
          overdue,
          progress: tasks.length ? Math.round((done / tasks.length) * 100) : 0,
        },
        milestones,
      };
    },
    enabled: !!projectId,
  });
}
