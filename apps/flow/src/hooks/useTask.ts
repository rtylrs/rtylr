import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useTask(id: string | undefined) {
  return useQuery({
    queryKey: ['task', id],
    queryFn: () => flowApi.getTaskDetail(id!),
    enabled: !!id,
  });
}
