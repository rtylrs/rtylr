import { useQuery } from "@tanstack/react-query";
import type { CaretChatSummary } from "../types/chat";
import { createMemoryClient } from "../lib/memoryClient";
import { normalizeChatTitle } from "../lib/chatTitles";

interface UseCaretSessionsParams {
  userId: number;
  apiBaseUrl: string;
  /** How many recent chats to load; the sidebar shows 20. */
  limit?: number;
  enabled?: boolean;
}

interface UseCaretSessionsResult {
  sessions: CaretChatSummary[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<unknown>;
}

export function useCaretSessions({
  userId,
  apiBaseUrl,
  limit = 20,
  enabled = true,
}: UseCaretSessionsParams): UseCaretSessionsResult {
  const { data, isLoading, error, refetch } = useQuery<CaretChatSummary[]>({
    // Keyed under ["caret", "sessions", userId] so chat mutations refresh every list.
    queryKey: ["caret", "sessions", userId, limit],
    queryFn: async (): Promise<CaretChatSummary[]> => {
      const res = await createMemoryClient(apiBaseUrl).listChats({
        limit,
        offset: 0,
      });
      return (res.chats ?? [])
        .filter((chat): chat is typeof chat & { id: string } =>
          Boolean(chat.id),
        )
        .map((chat) => ({
          id: chat.id,
          title: normalizeChatTitle(chat.title),
          summary: chat.summary || undefined,
          updatedAt: chat.updatedAt || new Date().toISOString(),
          pinned: chat.pinned,
          archived: chat.archived,
          projectId: chat.projectId || undefined,
        }));
    },
    enabled: enabled && userId > 0,
  });

  return {
    sessions: data ?? [],
    isLoading,
    error: error ?? null,
    refetch,
  };
}
