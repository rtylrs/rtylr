import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { flowApi } from '../../shared/services/flowApi';

/**
 * Get-or-creates the caller's personal "My Tasks" project and returns its ids.
 * Backs both the My Tasks landing redirect and the sidebar active-state.
 */
export function useMyTasksProject({ enabled = true } = {}) {
  return useQuery({
    queryKey: ['my-tasks-project'],
    enabled,
    queryFn: () => unwrap(flowApi.getMyTasksProject()),
    staleTime: 5 * 60 * 1000,
  });
}
