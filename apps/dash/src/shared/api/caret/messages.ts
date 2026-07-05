import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { safe } from '../../lib/api';

const messageSchema = z.object({
  id: z.string(),
  role: z.enum(['user', 'assistant', 'system', 'tool']),
  content: z.string().default(''),
  createdAt: z.string(),
  tokensUsed: z.number().optional(),
});

const messagesResponseSchema = z.object({
  messages: z.array(messageSchema),
});

export type StoredMessage = z.infer<typeof messageSchema>;

/** Fetch a single chat's full message history. */
export async function getChatMessages(chatId: string): Promise<StoredMessage[]> {
  const res = await unwrap(safe.get(`/api/hub/memory/v1/chats/${encodeURIComponent(chatId)}/messages`));
  const parsed = messagesResponseSchema.safeParse(res);
  if (parsed.success) return parsed.data.messages;
  return [];
}
