import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { supportService } from '../../shared/services/supportService';
import { conversationSchema } from '../../shared/schemas/conversations';

export function useConversations(params?: {
  status?: string;
  assignee?: string;
  channel?: string;
}) {
  return useQuery({
    queryKey: ['conversations', params],
    queryFn: async () => {
      const res = await unwrap(supportService.getConversations(params));
      const raw = res ?? [];
      return z.array(conversationSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useConversation(id: number) {
  return useQuery({
    queryKey: ['conversation', id],
    queryFn: async () => {
      const res = await unwrap(supportService.getConversation(id));
      return conversationSchema.parse(res);
    },
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
