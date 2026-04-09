import { useMemo } from 'react';
import { useFlowStore } from '../stores/flowStore';

export function useRoadmap() {
  const { projects, tasks } = useFlowStore();
  return useMemo(() => {
    return projects.map((p) => {
      const pTasks = tasks.filter((t) => t.projectId === p.id);
      const done = pTasks.filter((t) => t.status === 'done').length;
      return {
        ...p,
        tasks: pTasks,
        doneCount: done,
        progress: pTasks.length ? Math.round((done / pTasks.length) * 100) : 0,
      };
    });
  }, [projects, tasks]);
}
