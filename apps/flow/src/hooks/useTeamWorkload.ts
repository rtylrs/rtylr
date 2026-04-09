import { useMemo } from 'react';
import { useFlowStore } from '../stores/flowStore';

export function useTeamWorkload() {
  const { members, tasks } = useFlowStore();
  return useMemo(() => {
    return members.map((m) => {
      const mTasks = tasks.filter(
        (t) => t.assigneeId === m.id && t.status !== 'done' && t.status !== 'cancelled' && !t.archived
      );
      const assignedHours = mTasks.reduce((sum, t) => sum + (t.estimate || 4), 0);
      return {
        ...m,
        tasks: mTasks,
        assignedHours,
        load: m.capacity > 0 ? assignedHours / m.capacity : 0,
      };
    });
  }, [members, tasks]);
}
