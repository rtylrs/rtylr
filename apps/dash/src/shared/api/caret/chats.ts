import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { safe } from '../../lib/api';

const chatSummarySchema = z.object({
  id: z.string(),
  title: z.string(),
  summary: z.string().optional(),
  updatedAt: z.string(),
  pinned: z.boolean().optional(),
  archived: z.boolean().optional(),
});

const chatsResponseSchema = z.object({
  chats: z.array(chatSummarySchema),
});

export type ChatSummary = z.infer<typeof chatSummarySchema>;

/** Fetch the user's recent chats from the memory module. */
export async function listChats(): Promise<ChatSummary[]> {
  const res = await unwrap(safe.get('/api/hub/memory/v1/chats'));
  const parsed = chatsResponseSchema.safeParse(res);
  if (parsed.success) return parsed.data.chats;
  return [];
}

/** Delete a single chat session. */
export async function deleteChat(chatId: string): Promise<void> {
  await unwrap(safe.delete(`/api/hub/memory/v1/chats/${encodeURIComponent(chatId)}`));
}

/** Rename a chat session. */
export async function renameChat(chatId: string, title: string): Promise<void> {
  await unwrap(safe.put(`/api/hub/memory/v1/chats/${encodeURIComponent(chatId)}`, { title }));
}

/** Toggle the pinned flag on a chat. */
export async function setChatPinned(chatId: string, pinned: boolean): Promise<void> {
  await unwrap(safe.put(`/api/hub/memory/v1/chats/${encodeURIComponent(chatId)}/pin`, { pinned }));
}
