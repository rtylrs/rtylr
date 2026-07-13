import { useQuery } from '@tanstack/react-query';
import type { CaretChatSummary } from '../types/chat';
import { createMemoryClient } from '../lib/memoryClient';
import { normalizeChatTitle } from '../lib/chatTitles';

interface UseCaretSessionsParams {
  userId: number;
  apiBaseUrl: string;
}

interface UseCaretSessionsResult {
  sessions: CaretChatSummary[];
  isLoading: boolean;
}

export function useCaretSessions({
  userId,
  apiBaseUrl,
}: UseCaretSessionsParams): UseCaretSessionsResult {
  const { data, isLoading } = useQuery<CaretChatSummary[]>({
    queryKey: ['caret', 'sessions', userId],
    queryFn: async (): Promise<CaretChatSummary[]> => {
      const res = await createMemoryClient(apiBaseUrl).listChats({ limit: 20, offset: 0 });
      return (res.chats ?? [])
        .filter((chat): chat is typeof chat & { id: string } => Boolean(chat.id))
        .map((chat) => ({
          id: chat.id,
          title: normalizeChatTitle(chat.title),
          summary: chat.summary || undefined,
          updatedAt: chat.updatedAt || new Date().toISOString(),
          pinned: chat.pinned,
          archived: chat.archived,
        }));
    },
    enabled: userId > 0,
  });

  return {
    sessions: data ?? [],
    isLoading,
  };
}
