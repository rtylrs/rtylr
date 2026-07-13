import { useCallback } from 'react';
import type { CaretMessage } from '../types/message';
import { createMemoryClient } from '../lib/memoryClient';
import type { ChatMessage } from '@voxire/voxie/hub/memory/v1/types';

interface UseLoadCaretChatParams {
  userId: number;
  apiBaseUrl: string;
}

function toCaretMessage(m: ChatMessage & { content: string }, index: number): CaretMessage {
  const role = m.role === 'assistant' || m.role === 'system' ? m.role : 'user';
  return {
    id: `hist-${index}`,
    role,
    blocks: [{ id: `hist-${index}-text`, type: 'text', content: m.content }],
    createdAt: m.createdAt ?? new Date().toISOString(),
    content: m.content,
  };
}

/** Returns a loader that fetches a past session's messages from the server. */
export function useLoadCaretChat({
  userId: _userId,
  apiBaseUrl,
}: UseLoadCaretChatParams): (sessionId: string) => Promise<CaretMessage[]> {
  return useCallback(
    async (sessionId: string): Promise<CaretMessage[]> => {
      const res = await createMemoryClient(apiBaseUrl).getChatMessages({ sessionId });
      return (res.messages ?? [])
        .filter((m): m is ChatMessage & { content: string } => Boolean(m.content))
        .map(toCaretMessage);
    },
    [apiBaseUrl]
  );
}
