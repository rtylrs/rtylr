import { useMemo } from 'react';
import { useFlowStore } from '../stores/flowStore';

export function useMyIssues(userId: string = 'u1') {
  const { tasks } = useFlowStore();
  return useMemo(() => {
    return tasks.filter((t) => t.assigneeId === userId && !t.archived);
  }, [tasks, userId]);
}
