import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { useFlowStore } from '../../shared/stores/flowStore';
import { buildTeamWorkload } from '../../shared/lib/localFlowComputed';

export function useTeamWorkload() {
  const members = useFlowStore((s) => s.members);
  const tasks = useFlowStore((s) => s.tasks);
  const computed = useMemo(() => buildTeamWorkload(members, tasks), [members, tasks]);

  const q = useQuery({
    queryKey: ['team-workload'],
    queryFn: () => unwrap(flowApi.getTeamWorkload()),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? computed : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
