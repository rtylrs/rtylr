import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface CaretMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
}

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
  addMessage: (msg: CaretMessage) => void;
  appendToLastAssistantMessage: (token: string) => void;
  startAssistantMessage: () => void;
  setStreaming: (streaming: boolean) => void;
  reset: () => void;
}

export type CaretStore = CaretState & CaretActions;

const initialState: CaretState = {
  isOpen: false,
  sessionId: null,
  messages: [],
  isStreaming: false,
};

export const useCaretStore = create<CaretStore>()(
  persist(
    (set) => ({
      ...initialState,

      open() {
        set({ isOpen: true });
      },

      close() {
        set({ isOpen: false });
      },

      toggle() {
        set((state) => ({ isOpen: !state.isOpen }));
      },

      setSessionId(id: string) {
        set({ sessionId: id });
      },

      addMessage(msg: CaretMessage) {
        set((state) => ({ messages: [...state.messages, msg] }));
      },

      appendToLastAssistantMessage(token: string) {
        set((state) => {
          const messages = [...state.messages];
          const lastIndex = messages.length - 1;
          if (lastIndex >= 0 && messages[lastIndex].role === 'assistant') {
            messages[lastIndex] = {
              ...messages[lastIndex],
              content: messages[lastIndex].content + token,
            };
          }
          return { messages };
        });
      },

      startAssistantMessage() {
        set((state) => ({
          isStreaming: true,
          messages: [
            ...state.messages,
            {
              id: `assistant-${Date.now()}`,
              role: 'assistant' as const,
              content: '',
              createdAt: new Date().toISOString(),
            },
          ],
        }));
      },

      setStreaming(streaming: boolean) {
        set({ isStreaming: streaming });
      },

      reset() {
        set({ messages: [], sessionId: null, isStreaming: false });
      },
    }),
    {
      name: 'caret-session',
      partialize: (state) => ({ sessionId: state.sessionId }),
    }
  )
);
