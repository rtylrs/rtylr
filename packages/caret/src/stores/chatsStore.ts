import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { CaretChatSummary } from '../types/chat';

interface ChatsState {
  /** Active chat session id. */
  activeChatId: string | null;
  /** Locally-cached summaries for chats the user has touched recently. */
  recent: CaretChatSummary[];
}

interface ChatsActions {
  setActiveChat: (id: string | null) => void;
  upsertRecent: (chat: CaretChatSummary) => void;
  setRecent: (chats: CaretChatSummary[]) => void;
  removeRecent: (id: string) => void;
  clear: () => void;
}

export type ChatsStore = ChatsState & ChatsActions;

const MAX_RECENT = 20;

export const useChatsStore = create<ChatsStore>()(
  persist(
    (set) => ({
      activeChatId: null,
      recent: [],
      setActiveChat: (id) => set({ activeChatId: id }),
      upsertRecent: (chat) =>
        set((s) => {
          const filtered = s.recent.filter((c) => c.id !== chat.id);
          return { recent: [chat, ...filtered].slice(0, MAX_RECENT) };
        }),
      setRecent: (chats) => set({ recent: chats.slice(0, MAX_RECENT) }),
      removeRecent: (id) => set((s) => ({ recent: s.recent.filter((c) => c.id !== id) })),
      clear: () => set({ activeChatId: null, recent: [] }),
    }),
    { name: 'caret-chats' }
  )
);
