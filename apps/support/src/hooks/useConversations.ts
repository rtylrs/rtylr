import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { supportService } from '../services/supportService';

export function useConversations(params?: { status?: string; assignee?: string; channel?: string }) {
  return useQuery({
    queryKey: ['conversations', params],
    queryFn: () => supportService.getConversations(params).then((r) => r.data),
  });
}

export function useConversation(id: number) {
  return useQuery({
    queryKey: ['conversation', id],
    queryFn: () => supportService.getConversation(id).then((r) => r.data),
    enabled: !!id,
  });
}

export function useAssignConversation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, assigneeId }: { id: number; assigneeId: number }) =>
      supportService.assignConversation(id, assigneeId).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['conversations'] });
    },
  });
}
