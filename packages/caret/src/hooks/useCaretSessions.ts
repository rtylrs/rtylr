import { useQuery } from '@tanstack/react-query';
import type { CaretChatSummary } from '../types/chat';
import { authHeaders } from '../lib/auth';
import { normalizeChatTitle } from '../lib/chatTitles';

interface RawSession {
  id?: string;
  session_id?: string;
  title?: string;
  summary?: string;
  updatedAt?: string;
  updated_at?: string;
  pinned?: boolean;
  archived?: boolean;
}

interface SessionsResponse {
  sessions?: RawSession[];
  chats?: RawSession[];
}

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
  const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl : `${apiBaseUrl}/`;
  const url = `${baseUrl}memory/chats`;

  const { data, isLoading } = useQuery<CaretChatSummary[]>({
    queryKey: ['caret', 'sessions', userId],
    queryFn: async (): Promise<CaretChatSummary[]> => {
      const response = await fetch(url, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          ...authHeaders(),
        },
        body: JSON.stringify({ user_id: userId, limit: 20, offset: 0 }),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch sessions: ${response.statusText}`);
      }

      const json = (await response.json()) as SessionsResponse;
      const rows = json.chats?.length ? json.chats : (json.sessions ?? []);
      return rows
        .filter((s) => s.id || s.session_id)
        .map((s) => ({
          id: s.id ?? s.session_id ?? '',
          title: normalizeChatTitle(s.title),
          summary: s.summary,
          updatedAt: s.updatedAt ?? s.updated_at ?? new Date().toISOString(),
          pinned: s.pinned,
          archived: s.archived,
        }));
    },
    enabled: userId > 0,
  });

  return {
    sessions: data ?? [],
    isLoading,
  };
}
