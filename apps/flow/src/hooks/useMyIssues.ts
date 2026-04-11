import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';
import { isFlowLocal } from '../lib/flowLocalMode';
import { isTaskMine } from '../lib/currentUser';
import { useFlowStore } from '../stores/flowStore';

/** Tasks assigned to the current user (My tasks). */
export function useMyIssues() {
  const tasks = useFlowStore((s) => s.tasks);
  const localList = useMemo(
    () =>
      tasks.filter(
        (t) =>
          isTaskMine(t.assigneeId) &&
          !t.archived &&
          t.status !== 'done' &&
          t.status !== 'cancelled'
      ),
    [tasks]
  );
  const q = useQuery({
    queryKey: ['my-tasks'],
    queryFn: () => flowApi.getMyTasks(),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? localList : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
