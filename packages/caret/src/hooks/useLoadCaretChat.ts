import { useCallback } from 'react';
import type { CaretMessage } from '../types/message';
import { authHeaders } from '../lib/auth';

interface RawMessage {
  id: number;
  role: string;
  content: string;
  created_at: string;
}

interface MessagesResponse {
  context?: {
    messages?: RawMessage[];
  };
}

interface UseLoadCaretChatParams {
  userId: number;
  apiBaseUrl: string;
}

function toCaretMessage(m: RawMessage, index: number): CaretMessage {
  const role = m.role === 'assistant' || m.role === 'system' ? m.role : 'user';
  return {
    id: `hist-${index}`,
    role,
    blocks: [{ id: `hist-${index}-text`, type: 'text', content: m.content }],
    createdAt: m.created_at,
    content: m.content,
  };
}

/** Returns a loader that fetches a past session's messages from the server. */
export function useLoadCaretChat({
  userId,
  apiBaseUrl,
}: UseLoadCaretChatParams): (sessionId: string) => Promise<CaretMessage[]> {
  const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl : `${apiBaseUrl}/`;

  return useCallback(
    async (sessionId: string): Promise<CaretMessage[]> => {
      const url = `${baseUrl}memory/chats/${encodeURIComponent(sessionId)}/messages`;
      const response = await fetch(url, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          ...authHeaders(),
        },
        body: JSON.stringify({ session_id: sessionId, user_id: userId }),
      });

      if (!response.ok) {
        throw new Error(`Failed to load chat: ${response.statusText}`);
      }

      const json = (await response.json()) as MessagesResponse;
      const messages = json.context?.messages ?? [];
      return messages.filter((m) => m.content).map(toCaretMessage);
    },
    [baseUrl, userId]
  );
}
