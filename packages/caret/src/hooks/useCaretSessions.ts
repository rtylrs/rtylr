import { useQuery } from '@tanstack/react-query';

export interface CaretSession {
  id: string;
  title?: string;
  createdAt: string;
  updatedAt: string;
}

interface SessionsResponse {
  sessions: CaretSession[];
}

interface UseCaretSessionsParams {
  userId: number;
  apiBaseUrl: string;
}

interface UseCaretSessionsResult {
  sessions: CaretSession[];
  isLoading: boolean;
}

export function useCaretSessions({
  userId,
  apiBaseUrl,
}: UseCaretSessionsParams): UseCaretSessionsResult {
  const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl : `${apiBaseUrl}/`;
  const url = `${baseUrl}memory/chats`;

  const { data, isLoading } = useQuery<SessionsResponse>({
    queryKey: ['caret', 'sessions', userId],
    queryFn: async (): Promise<SessionsResponse> => {
      const response = await fetch(url, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ user_id: userId, limit: 20, offset: 0 }),
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch sessions: ${response.statusText}`);
      }

      return response.json() as Promise<SessionsResponse>;
    },
    enabled: userId > 0,
  });

  return {
    sessions: data?.sessions ?? [],
    isLoading,
  };
}
