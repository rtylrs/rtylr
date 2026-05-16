import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import {
  deleteChat,
  listChats,
  renameChat,
  setChatPinned,
  type ChatSummary,
} from '../../api/caret';

/** React Query hook for the user's chat history list. */
export const useCaretChats = () =>
  useQuery<ChatSummary[]>({
    queryKey: ['caret', 'chats'],
    queryFn: listChats,
    staleTime: 30_000,
  });

export const useDeleteCaretChat = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (chatId: string) => deleteChat(chatId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['caret', 'chats'] });
      toast.success('Chat deleted');
    },
    onError: () => toast.error('Failed to delete chat'),
  });
};

export const useRenameCaretChat = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ chatId, title }: { chatId: string; title: string }) => renameChat(chatId, title),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['caret', 'chats'] }),
    onError: () => toast.error('Failed to rename chat'),
  });
};

export const useSetCaretChatPinned = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ chatId, pinned }: { chatId: string; pinned: boolean }) =>
      setChatPinned(chatId, pinned),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['caret', 'chats'] }),
    onError: () => toast.error('Failed to pin chat'),
  });
};
