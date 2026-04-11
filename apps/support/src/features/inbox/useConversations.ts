import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { supportService } from '../../shared/services/supportService';

export function useConversations(params?: {
  status?: string;
  assignee?: string;
  channel?: string;
}) {
  return useQuery({
    queryKey: ['conversations', params],
    queryFn: () => unwrap(supportService.getConversations(params)),
  });
}

export function useConversation(id: number) {
  return useQuery({
    queryKey: ['conversation', id],
    queryFn: () => unwrap(supportService.getConversation(id)),
    enabled: !!id,
  });
}

export function useAssignConversation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, assigneeId }: { id: number; assigneeId: number }) =>
      unwrap(supportService.assignConversation(id, assigneeId)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['conversations'] });
    },
  });
}
