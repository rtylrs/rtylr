import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';
import { isFlowLocal } from '../lib/flowLocalMode';
import { useFlowStore } from '../stores/flowStore';

export function useProjectOverview(projectId: string | undefined) {
  const tasks = useFlowStore((s) => s.tasks);
  const milestones = useFlowStore((s) => s.milestones);

  const localData = useMemo(() => {
    if (!projectId) return null;
    const list = tasks.filter((t) => t.projectId === projectId);
    const done = list.filter((t) => t.status === 'done').length;
    const inProgress = list.filter(
      (t) => t.status === 'in_progress' || t.status === 'in_review'
    ).length;
    const overdue = list.filter((t) => {
      if (!t.dueDate || t.status === 'done') return false;
      return new Date(t.dueDate) < new Date();
    }).length;
    const ms = milestones.filter((m) => m.projectId === projectId);
    return {
      tasks: list as any,
      stats: {
        total: list.length,
        done,
        inProgress,
        overdue,
        progress: list.length ? Math.round((done / list.length) * 100) : 0,
      },
      milestones: ms as any,
    };
  }, [projectId, tasks, milestones]);

  const q = useQuery({
    queryKey: ['project-overview', projectId],
    queryFn: async () => {
      if (!projectId) return null;
      const [taskList, ms] = await Promise.all([
        flowApi.getTasks({ projectId }),
        flowApi.getMilestones(projectId),
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
    enabled: !isFlowLocal() && !!projectId,
  });
  return {
    ...q,
    data: isFlowLocal() ? localData : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
