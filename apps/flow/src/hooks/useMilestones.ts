import { useMemo } from 'react';
import { useFlowStore } from '../stores/flowStore';

export function useMilestones(projectId?: string) {
  const { milestones, tasks, projects } = useFlowStore();
  return useMemo(() => {
    const filtered = projectId
      ? milestones.filter((m) => m.projectId === projectId)
      : milestones;
    return filtered.map((m) => {
      const mTasks = tasks.filter((t) => t.milestoneId === m.id);
      const done = mTasks.filter((t) => t.status === 'done').length;
      const project = projects.find((p) => p.id === m.projectId);
      return {
        ...m,
        project,
        taskCount: mTasks.length,
        doneCount: done,
        computedProgress: mTasks.length
          ? Math.round((done / mTasks.length) * 100)
          : m.progress,
      };
    });
  }, [milestones, tasks, projects, projectId]);
}
