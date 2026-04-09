import { useMemo } from 'react';
import { useFlowStore } from '../stores/flowStore';

export function useCycles(workspaceId?: string) {
  const { cycles, tasks } = useFlowStore();
  return useMemo(() => {
    const ws = workspaceId;
    const filtered = ws ? cycles.filter((c) => c.workspaceId === ws) : cycles;
    return filtered.map((c) => {
      const cTasks = tasks.filter((t) => t.cycleId === c.id);
      const done = cTasks.filter((t) => t.status === 'done').length;
      return {
        ...c,
        scope: cTasks.length || c.scope,
        completed: done || c.completed,
        tasks: cTasks,
      };
    });
  }, [cycles, tasks, workspaceId]);
}

export function useCycleById(id: string | undefined) {
  const { cycles, tasks } = useFlowStore();
  return useMemo(() => {
    if (!id) return null;
    const cycle = cycles.find((c) => c.id === id);
    if (!cycle) return null;
    const cycleTasks = tasks.filter((t) => t.cycleId === id);
    const done = cycleTasks.filter((t) => t.status === 'done').length;
    return {
      ...cycle,
      tasks: cycleTasks,
      scope: cycleTasks.length,
      completed: done,
    };
  }, [cycles, tasks, id]);
}
