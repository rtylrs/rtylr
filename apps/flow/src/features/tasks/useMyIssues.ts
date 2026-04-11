import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { isTaskMine } from '../../shared/lib/currentUser';
import { useFlowStore } from '../../shared/stores/flowStore';

/** Tasks assigned to the current user (My tasks). */
export function useMyIssues() {
  const tasks = useFlowStore((s) => s.tasks);
  const localList = useMemo(
    () =>
      tasks.filter(
        (t) =>
          isTaskMine(t.assigneeId) && !t.archived && t.status !== 'done' && t.status !== 'cancelled'
      ),
    [tasks]
  );
  const q = useQuery({
    queryKey: ['my-tasks'],
    queryFn: () => unwrap(flowApi.getMyTasks()),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? localList : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
