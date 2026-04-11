import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';
import { isFlowLocal } from '../lib/flowLocalMode';
import { useFlowStore } from '../stores/flowStore';

export function useTask(id: string | undefined) {
  const fromStore = useFlowStore((s) => (id ? s.tasks.find((t) => t.id === id) : undefined));
  const q = useQuery({
    queryKey: ['task', id],
    queryFn: () => flowApi.getTaskDetail(id!),
    enabled: !isFlowLocal() && !!id,
  });
  return {
    ...q,
    data: isFlowLocal() ? fromStore : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
