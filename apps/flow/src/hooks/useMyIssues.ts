import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useMyIssues() {
  return useQuery({
    queryKey: ['my-tasks'],
    queryFn: () => flowApi.getMyTasks(),
  });
}
