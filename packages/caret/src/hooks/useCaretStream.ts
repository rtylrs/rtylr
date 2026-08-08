import { useCallback, useRef } from 'react';
import { useCaretStore } from '../stores/caretStore';
import { useCaretSettingsStore } from '../stores/settingsStore';
import { consumeSse } from '../protocol/sse';
import type { CaretStreamEvent } from '../protocol/events';
import { readAuthToken } from '../lib/auth';
import type { CaretPageContext } from '../types/context';
import type { CaretMode } from '../types/mode';

export interface UseCaretStreamParams {
  userId: number;
  apiBaseUrl: string;
  systemContext?: string;
  pageContext?: CaretPageContext;
}

export interface UseCaretStreamResult {
  sendMessage: (prompt: string, modeOverride?: CaretMode) => Promise<void>;
  stopStreaming: (tabId?: string) => void;
}

interface OpenThinking {
  blockId: string;
  startedAt: number;
}

export function useCaretStream({
  userId,
  apiBaseUrl,
  systemContext,
  pageContext,
}: UseCaretStreamParams): UseCaretStreamResult {
  const abortRef = useRef<Map<string, AbortController>>(new Map());

  const stopStreaming = useCallback((tabId?: string): void => {
    const store = useCaretStore.getState();
    const targetTabId = tabId ?? store.activeTabId;
    const controller = abortRef.current.get(targetTabId);
    controller?.abort();
    if (abortRef.current.get(targetTabId) === controller) {
      abortRef.current.delete(targetTabId);
    }
    store.setStreaming(false, targetTabId);
  }, []);

  const sendMessage = useCallback(
    async (prompt: string, modeOverride?: CaretMode): Promise<void> => {
      const store = useCaretStore.getState();
      const tab = store.tabs.find((candidate) => candidate.id === store.activeTabId);
      if (!prompt.trim() || !tab || tab.isStreaming) return;

      const tabId = tab.id;

      const userMessageId = `user-${Date.now()}`;
      store.addMessage(
        {
          id: userMessageId,
          role: 'user',
          blocks: [],
          content: prompt.trim(),
          createdAt: new Date().toISOString(),
        },
        tabId
      );

      store.startAssistantMessage(tabId);
      const assistantMessageId = `assistant-${Date.now()}`;

      const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl : `${apiBaseUrl}/`;
      const url = `${baseUrl}ai/generate-text-stream`;
      const controller = new AbortController();
      abortRef.current.set(tabId, controller);

      const openThinking: Record<string, OpenThinking> = {};

      try {
        const token = readAuthToken();
        const mode = modeOverride ?? useCaretSettingsStore.getState().mode;
        const response = await fetch(url, {
          method: 'POST',
          credentials: 'include',
          signal: controller.signal,
          headers: {
            'Content-Type': 'application/json',
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          body: JSON.stringify({
            prompt: prompt.trim(),
            session_id:
              useCaretStore.getState().tabs.find((candidate) => candidate.id === tabId)
                ?.sessionId ?? '',
            user_id: userId,
            system_context: systemContext ?? '',
            page_context: pageContext,
            mode,
          }),
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        for await (const event of consumeSse(response, assistantMessageId)) {
          dispatch(event, openThinking, tabId);
        }

        useCaretStore.getState().setStreaming(false, tabId);
      } catch (error) {
        if ((error as Error).name === 'AbortError') return;
        useCaretStore.getState().setStreaming(false, tabId);
        const message =
          error instanceof Error
            ? error.message
            : 'An unexpected error occurred. Please try again.';
        useCaretStore.getState().appendBlockToLastAssistant(
          {
            id: `error-${Date.now()}`,
            type: 'text',
            content: `\n\n_Error: ${message}_`,
          },
          tabId
        );
      } finally {
        if (abortRef.current.get(tabId) === controller) {
          abortRef.current.delete(tabId);
          useCaretStore.getState().setStreaming(false, tabId);
        }
      }
    },
    [apiBaseUrl, userId, systemContext, pageContext]
  );

  /** Translate one server event into store mutations. */
  function dispatch(
    event: CaretStreamEvent,
    openThinking: Record<string, OpenThinking>,
    tabId: string
  ): void {
    const store = useCaretStore.getState();
    switch (event.type) {
      case 'session':
        if (!store.tabs.find((candidate) => candidate.id === tabId)?.sessionId) {
          store.setSessionId(event.sessionId, tabId);
        }
        break;

      case 'text_delta':
        store.appendToLastAssistantMessage(event.delta, tabId);
        break;

      case 'thinking_start':
        openThinking[event.blockId] = { blockId: event.blockId, startedAt: Date.now() };
        store.appendBlockToLastAssistant(
          {
            id: event.blockId,
            type: 'thinking',
            content: '',
          },
          tabId
        );
        break;

      case 'thinking_delta':
        store.updateBlockOnLastAssistant(
          event.blockId,
          (b) => (b.type === 'thinking' ? { ...b, content: b.content + event.delta } : b),
          tabId
        );
        break;

      case 'thinking_end': {
        const opened = openThinking[event.blockId];
        const duration = event.durationMs ?? (opened ? Date.now() - opened.startedAt : undefined);
        store.updateBlockOnLastAssistant(
          event.blockId,
          (b) => (b.type === 'thinking' ? { ...b, durationMs: duration } : b),
          tabId
        );
        delete openThinking[event.blockId];
        break;
      }

      case 'tool_call':
        store.appendBlockToLastAssistant(
          {
            id: event.blockId,
            type: 'tool_call',
            name: event.name,
            mode: event.mode,
            input: event.input,
            status: event.requiresApproval ? 'pending' : 'running',
            preview: event.preview,
            requiresApproval: event.requiresApproval,
          },
          tabId
        );
        break;

      case 'tool_result':
        store.setToolStatus(
          event.blockId,
          event.status,
          {
            output: event.output,
            summary: event.summary,
            error: event.error,
          },
          tabId
        );
        break;

      case 'message_end':
      case 'done':
        store.setStreaming(false, tabId);
        break;

      case 'error':
        store.appendBlockToLastAssistant(
          {
            id: `error-${Date.now()}`,
            type: 'text',
            content: `\n\n_${event.error}_`,
          },
          tabId
        );
        if (!event.recoverable) store.setStreaming(false, tabId);
        break;
    }
  }

  return { sendMessage, stopStreaming };
}
