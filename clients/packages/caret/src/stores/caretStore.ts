import { create } from "zustand";
import { persist } from "zustand/middleware";
import type {
  CaretMessage,
  ContentBlock,
  ToolCallBlock,
} from "../types/message";
import { normalizeChatTitle, titleFromUserMessage } from "../lib/chatTitles";

export type { CaretMessage } from "../types/message";

const NEW_CHAT_TITLE = "New chat";

export interface CaretTab {
  id: string;
  sessionId: string | null;
  title: string;
  projectId: string | null;
  messages: CaretMessage[];
  isStreaming: boolean;
  isLoading: boolean;
  isLoaded: boolean;
}

interface CaretState {
  isOpen: boolean;
  tabs: CaretTab[];
  activeTabId: string;
  /** Compatibility projections for older embedded Caret consumers. */
  messages: CaretMessage[];
  isStreaming: boolean;
}

export interface NewCaretTabInput {
  sessionId?: string | null;
  title?: string | null;
  projectId?: string | null;
}

interface CaretActions {
  open: () => void;
  close: () => void;
  toggle: () => void;
  newTab: (input?: NewCaretTabInput) => string;
  /** Focus the tab already showing a saved chat, or open one for it. */
  openChat: (input: { sessionId: string; title?: string; projectId?: string }) => string;
  setActiveTab: (id: string) => void;
  closeTab: (id: string) => void;
  setTabTitle: (title: string, tabId?: string) => void;
  setTabLoading: (loading: boolean, tabId?: string) => void;
  setTabLoaded: (loaded: boolean, tabId?: string) => void;
  setSessionId: (id: string | null, tabId?: string) => void;
  setMessages: (messages: CaretMessage[], tabId?: string) => void;
  addMessage: (msg: CaretMessage, tabId?: string) => void;
  appendToLastAssistantMessage: (token: string, tabId?: string) => void;
  startAssistantMessage: (tabId?: string) => void;
  /** Append a block to the last assistant message. Creates one if missing. */
  appendBlockToLastAssistant: (block: ContentBlock, tabId?: string) => void;
  /** Record how long the last assistant reply took. */
  setLastAssistantResponseTime: (responseMs: number, tabId?: string) => void;
  setLastAssistantStoredId: (storedId: string, tabId?: string) => void;
  /** Attach suggested follow-up questions to the last assistant reply. */
  setLastAssistantFollowUps: (items: string[], tabId?: string) => void;
  /** Update an existing block on the last assistant message. */
  updateBlockOnLastAssistant: (
    blockId: string,
    patch: (block: ContentBlock) => ContentBlock,
    tabId?: string,
  ) => void;
  setStreaming: (streaming: boolean, tabId?: string) => void;
  setToolStatus: (
    blockId: string,
    status: ToolCallBlock["status"],
    extras?: Partial<ToolCallBlock>,
    tabId?: string,
  ) => void;
  reset: (tabId?: string) => void;
}

export type CaretStore = CaretState & CaretActions;

interface PersistedCaretTab {
  id: string;
  sessionId: string | null;
  title: string;
  projectId?: string | null;
}

interface PersistedCaretState {
  isOpen?: boolean;
  activeTabId?: string;
  tabs?: PersistedCaretTab[];
}

function createId(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
}

function createTab(input: NewCaretTabInput & { id?: string } = {}): CaretTab {
  const sessionId = input.sessionId ?? null;
  return {
    id: input.id ?? createId("tab"),
    sessionId,
    title: normalizeChatTitle(input.title) || NEW_CHAT_TITLE,
    projectId: input.projectId ?? null,
    messages: [],
    isStreaming: false,
    isLoading: false,
    isLoaded: !sessionId,
  };
}

const firstTab = createTab();
const EMPTY_MESSAGES: CaretMessage[] = [];

const initialState: CaretState = {
  isOpen: false,
  tabs: [firstTab],
  activeTabId: firstTab.id,
  messages: firstTab.messages,
  isStreaming: firstTab.isStreaming,
};

function activeTabId(state: CaretState, tabId?: string): string {
  return tabId ?? state.activeTabId;
}

function updateTab(
  state: CaretState,
  tabId: string | undefined,
  updater: (tab: CaretTab) => CaretTab,
): Partial<CaretState> {
  const id = activeTabId(state, tabId);
  const index = state.tabs.findIndex((tab) => tab.id === id);
  if (index < 0) return {};

  const nextTab = updater(state.tabs[index]);
  if (nextTab === state.tabs[index]) return {};

  const tabs = [...state.tabs];
  tabs[index] = nextTab;
  const activeTab = tabs.find((tab) => tab.id === state.activeTabId);
  return {
    tabs,
    messages: activeTab?.messages ?? EMPTY_MESSAGES,
    isStreaming: activeTab?.isStreaming ?? false,
  };
}

function titleFromMessages(messages: CaretMessage[]): string | null {
  const firstUserMessage = messages.find(
    (message) => message.role === "user" && message.content,
  );
  return firstUserMessage?.content
    ? titleFromUserMessage(firstUserMessage.content)
    : null;
}

function mergePersistedState(
  persistedState: unknown,
  currentState: CaretStore,
): CaretStore {
  const persisted = (persistedState ?? {}) as PersistedCaretState;
  const persistedTabs = Array.isArray(persisted.tabs)
    ? persisted.tabs
        .filter((tab): tab is PersistedCaretTab => Boolean(tab?.id))
        .map((tab) =>
          createTab({
            id: tab.id,
            sessionId: tab.sessionId,
            title: tab.title,
            projectId: tab.projectId,
          }),
        )
    : [];
  const tabs = persistedTabs.length > 0 ? persistedTabs : currentState.tabs;
  const activeTabId = tabs.some((tab) => tab.id === persisted.activeTabId)
    ? persisted.activeTabId!
    : tabs[0].id;
  const activeTab = tabs.find((tab) => tab.id === activeTabId) ?? tabs[0];

  return {
    ...currentState,
    isOpen:
      typeof persisted.isOpen === "boolean"
        ? persisted.isOpen
        : currentState.isOpen,
    tabs,
    activeTabId,
    messages: activeTab.messages,
    isStreaming: activeTab.isStreaming,
  };
}

const emptyAssistant = (): CaretMessage => ({
  id: createId("assistant"),
  role: "assistant",
  blocks: [],
  content: "",
  createdAt: new Date().toISOString(),
});

export const useCaretStore = create<CaretStore>()(
  persist(
    (set, get) => ({
      ...initialState,

      open: () => set({ isOpen: true }),
      close: () => set({ isOpen: false }),
      toggle: () => set((state) => ({ isOpen: !state.isOpen })),

      newTab: (input = {}) => {
        const tab = createTab(input);
        set((state) => ({
          tabs: [...state.tabs, tab],
          activeTabId: tab.id,
          messages: tab.messages,
          isStreaming: tab.isStreaming,
        }));
        return tab.id;
      },

      openChat: ({ sessionId, title, projectId }) => {
        const existing = get().tabs.find((tab) => tab.sessionId === sessionId);
        if (existing) {
          get().setActiveTab(existing.id);
          return existing.id;
        }
        return get().newTab({ sessionId, title, projectId });
      },

      setActiveTab: (id) =>
        set((state) => {
          const tab = state.tabs.find((candidate) => candidate.id === id);
          return tab
            ? {
                activeTabId: id,
                messages: tab.messages,
                isStreaming: tab.isStreaming,
              }
            : {};
        }),

      closeTab: (id) =>
        set((state) => {
          const index = state.tabs.findIndex((tab) => tab.id === id);
          if (index < 0) return {};

          if (state.tabs.length === 1) {
            const replacement = createTab({ id });
            return {
              tabs: [replacement],
              activeTabId: replacement.id,
              messages: replacement.messages,
              isStreaming: replacement.isStreaming,
            };
          }

          const tabs = state.tabs.filter((tab) => tab.id !== id);
          if (state.activeTabId !== id) return { tabs };

          const nextIndex = Math.min(index, tabs.length - 1);
          const nextTab = tabs[nextIndex];
          return {
            tabs,
            activeTabId: nextTab.id,
            messages: nextTab.messages,
            isStreaming: nextTab.isStreaming,
          };
        }),

      setTabTitle: (title, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({
            ...tab,
            title: normalizeChatTitle(title),
          })),
        ),

      setTabLoading: (isLoading, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({ ...tab, isLoading })),
        ),

      setTabLoaded: (isLoaded, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({ ...tab, isLoaded })),
        ),

      setSessionId: (id, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({ ...tab, sessionId: id })),
        ),

      setMessages: (messages, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({
            ...tab,
            messages,
            isLoaded: true,
            title:
              tab.title === NEW_CHAT_TITLE
                ? (titleFromMessages(messages) ?? tab.title)
                : tab.title,
          })),
        ),

      addMessage: (msg, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({
            ...tab,
            messages: [...tab.messages, msg],
            title:
              tab.title === NEW_CHAT_TITLE && msg.role === "user" && msg.content
                ? titleFromUserMessage(msg.content)
                : tab.title,
          })),
        ),

      appendToLastAssistantMessage: (token, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => {
            const messages = [...tab.messages];
            const lastIndex = messages.length - 1;
            if (lastIndex < 0 || messages[lastIndex].role !== "assistant")
              return tab;

            const last = messages[lastIndex];
            // Text continues the trailing text block; text after a tool call or
            // thought starts a new block so the reply reads in the order it ran.
            const blocks = [...last.blocks];
            const tail = blocks[blocks.length - 1];
            if (tail?.type === "text") {
              blocks[blocks.length - 1] = {
                ...tail,
                content: tail.content + token,
              };
            } else {
              blocks.push({
                id: `${last.id}-text-${blocks.length}`,
                type: "text",
                content: token,
              });
            }
            messages[lastIndex] = {
              ...last,
              blocks,
              content: (last.content ?? "") + token,
            };
            return { ...tab, messages };
          }),
        ),

      startAssistantMessage: (tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({
            ...tab,
            isStreaming: true,
            messages: [...tab.messages, emptyAssistant()],
          })),
        ),

      setLastAssistantFollowUps: (items, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => {
            const index = tab.messages.length - 1;
            if (index < 0 || tab.messages[index].role !== "assistant") return tab;
            const messages = [...tab.messages];
            messages[index] = { ...messages[index], followUps: items };
            return { ...tab, messages };
          }),
        ),

      setLastAssistantStoredId: (storedId, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => {
            const lastIndex = tab.messages.length - 1;
            if (lastIndex < 0 || tab.messages[lastIndex].role !== "assistant")
              return tab;
            const messages = [...tab.messages];
            messages[lastIndex] = { ...messages[lastIndex], storedId };
            return { ...tab, messages };
          }),
        ),

      setLastAssistantResponseTime: (responseMs, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => {
            const lastIndex = tab.messages.length - 1;
            if (lastIndex < 0 || tab.messages[lastIndex].role !== "assistant")
              return tab;
            const messages = [...tab.messages];
            messages[lastIndex] = { ...messages[lastIndex], responseMs };
            return { ...tab, messages };
          }),
        ),

      appendBlockToLastAssistant: (block, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => {
            const messages = [...tab.messages];
            let lastIndex = messages.length - 1;
            if (lastIndex < 0 || messages[lastIndex].role !== "assistant") {
              messages.push(emptyAssistant());
              lastIndex = messages.length - 1;
            }
            const last = messages[lastIndex];
            messages[lastIndex] = {
              ...last,
              blocks: [...last.blocks, block],
              content:
                block.type === "text"
                  ? [last.content ?? "", block.content]
                      .filter(Boolean)
                      .join("\n")
                  : last.content,
            };
            return { ...tab, messages };
          }),
        ),

      updateBlockOnLastAssistant: (blockId, patch, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => {
            const messages = [...tab.messages];
            const lastIndex = messages.length - 1;
            if (lastIndex < 0 || messages[lastIndex].role !== "assistant")
              return tab;
            const last = messages[lastIndex];
            messages[lastIndex] = {
              ...last,
              blocks: last.blocks.map((block) =>
                block.id === blockId ? patch(block) : block,
              ),
            };
            return { ...tab, messages };
          }),
        ),

      setStreaming: (isStreaming, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({ ...tab, isStreaming })),
        ),

      setToolStatus: (blockId, status, extras, tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({
            ...tab,
            messages: tab.messages.map((message) =>
              message.role !== "assistant"
                ? message
                : {
                    ...message,
                    blocks: message.blocks.map((block) =>
                      block.id === blockId && block.type === "tool_call"
                        ? { ...block, status, ...(extras ?? {}) }
                        : block,
                    ),
                  },
            ),
          })),
        ),

      reset: (tabId) =>
        set((state) =>
          updateTab(state, tabId, (tab) => ({
            ...createTab({ id: tab.id }),
            title: NEW_CHAT_TITLE,
          })),
        ),
    }),
    {
      name: "caret-session",
      version: 2,
      partialize: (state): PersistedCaretState => ({
        isOpen: state.isOpen,
        activeTabId: state.activeTabId,
        tabs: state.tabs.map(({ id, sessionId, title, projectId }) => ({
          id,
          sessionId,
          title,
          projectId,
        })),
      }),
      migrate: (persistedState) => persistedState as PersistedCaretState,
      merge: mergePersistedState,
    },
  ),
);
