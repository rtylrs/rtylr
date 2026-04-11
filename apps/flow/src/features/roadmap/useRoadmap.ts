import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { useFlowStore } from '../../shared/stores/flowStore';

export function useRoadmap() {
  const projects = useFlowStore((s) => s.projects);
  const tasks = useFlowStore((s) => s.tasks);
  const localData = useMemo(
    () =>
      projects.map((p) => {
        const pt = tasks.filter((t) => t.projectId === p.id);
        const doneCount = pt.filter((t) => t.status === 'done').length;
        return {
          id: p.id,
          workspaceId: p.workspaceId,
          name: p.name,
          color: p.color,
          description: p.description ?? '',
          memberAvatars: p.memberAvatars,
          taskCount: pt.length,
          tasks: pt as any,
          doneCount,
          progress: p.progress,
          startDate: p.startDate,
          targetDate: p.targetDate,
        };
      }),
    [projects, tasks]
  );

  const q = useQuery({
    queryKey: ['roadmap'],
    queryFn: () => unwrap(flowApi.getRoadmap()),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? localData : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
