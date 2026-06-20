import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { CaretMessage, ContentBlock, ToolCallBlock } from '../types/message';

export type { CaretMessage } from '../types/message';

interface CaretState {
  isOpen: boolean;
  sessionId: string | null;
  messages: CaretMessage[];
  isStreaming: boolean;
}

interface CaretActions {
  open: () => void;
  close: () => void;
  toggle: () => void;
  setSessionId: (id: string) => void;
  setMessages: (messages: CaretMessage[]) => void;
  addMessage: (msg: CaretMessage) => void;
  appendToLastAssistantMessage: (token: string) => void;
  startAssistantMessage: () => void;
  /** Append a block to the last assistant message. Creates one if missing. */
  appendBlockToLastAssistant: (block: ContentBlock) => void;
  /** Update an existing block on the last assistant message. */
  updateBlockOnLastAssistant: (
    blockId: string,
    patch: (block: ContentBlock) => ContentBlock
  ) => void;
  setStreaming: (streaming: boolean) => void;
  setToolStatus: (
    blockId: string,
    status: ToolCallBlock['status'],
    extras?: Partial<ToolCallBlock>
  ) => void;
  reset: () => void;
}

export type CaretStore = CaretState & CaretActions;

const initialState: CaretState = {
  isOpen: false,
  sessionId: null,
  messages: [],
  isStreaming: false,
};

const emptyAssistant = (): CaretMessage => ({
  id: `assistant-${Date.now()}`,
  role: 'assistant',
  blocks: [],
  content: '',
  createdAt: new Date().toISOString(),
});

export const useCaretStore = create<CaretStore>()(
  persist(
    (set) => ({
      ...initialState,

      open: () => set({ isOpen: true }),
      close: () => set({ isOpen: false }),
      toggle: () => set((s) => ({ isOpen: !s.isOpen })),
      setSessionId: (id) => set({ sessionId: id }),

      setMessages: (messages) => set({ messages }),

      addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),

      appendToLastAssistantMessage: (token) =>
        set((s) => {
          const messages = [...s.messages];
          const lastIndex = messages.length - 1;
          if (lastIndex < 0 || messages[lastIndex].role !== 'assistant') return { messages };

          const last = messages[lastIndex];
          // Maintain legacy `content` for clipboard/SSR fallbacks while we
          // transition the renderer to blocks.
          const blocks = last.blocks.length > 0 ? [...last.blocks] : [];
          let firstText = blocks.findIndex((b) => b.type === 'text');
          if (firstText === -1) {
            blocks.push({ id: `${last.id}-text-0`, type: 'text', content: token });
            firstText = blocks.length - 1;
          } else {
            const existing = blocks[firstText];
            if (existing.type === 'text') {
              blocks[firstText] = { ...existing, content: existing.content + token };
            }
          }
          messages[lastIndex] = {
            ...last,
            blocks,
            content: (last.content ?? '') + token,
          };
          return { messages };
        }),

      startAssistantMessage: () =>
        set((s) => ({
          isStreaming: true,
          messages: [...s.messages, emptyAssistant()],
        })),

      appendBlockToLastAssistant: (block) =>
        set((s) => {
          const messages = [...s.messages];
          let lastIndex = messages.length - 1;
          if (lastIndex < 0 || messages[lastIndex].role !== 'assistant') {
            messages.push(emptyAssistant());
            lastIndex = messages.length - 1;
          }
          const last = messages[lastIndex];
          messages[lastIndex] = {
            ...last,
            blocks: [...last.blocks, block],
            content:
              block.type === 'text'
                ? [last.content ?? '', block.content].filter(Boolean).join('\n')
                : last.content,
          };
          return { messages };
        }),

      updateBlockOnLastAssistant: (blockId, patch) =>
        set((s) => {
          const messages = [...s.messages];
          const lastIndex = messages.length - 1;
          if (lastIndex < 0 || messages[lastIndex].role !== 'assistant') return { messages };
          const last = messages[lastIndex];
          messages[lastIndex] = {
            ...last,
            blocks: last.blocks.map((b) => (b.id === blockId ? patch(b) : b)),
          };
          return { messages };
        }),

      setStreaming: (streaming) => set({ isStreaming: streaming }),

      setToolStatus: (blockId, status, extras) =>
        set((s) => {
          const messages = s.messages.map((m) =>
            m.role !== 'assistant'
              ? m
              : {
                  ...m,
                  blocks: m.blocks.map((b) =>
                    b.id === blockId && b.type === 'tool_call'
                      ? { ...b, status, ...(extras ?? {}) }
                      : b
                  ),
                }
          );
          return { messages };
        }),

      reset: () => set({ messages: [], sessionId: null, isStreaming: false }),
    }),
    {
      name: 'caret-session',
      partialize: (state) => ({ isOpen: state.isOpen }),
    }
  )
);
