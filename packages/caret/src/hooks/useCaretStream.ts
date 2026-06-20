import { useCallback, useRef } from 'react';
import { useCaretStore } from '../stores/caretStore';
import { useCaretSettingsStore } from '../stores/settingsStore';
import { consumeSse } from '../protocol/sse';
import type { CaretStreamEvent } from '../protocol/events';
import { readAuthToken } from '../lib/auth';
import type { CaretPageContext } from '../types/context';

export interface UseCaretStreamParams {
  userId: number;
  apiBaseUrl: string;
  systemContext?: string;
  pageContext?: CaretPageContext;
}

export interface UseCaretStreamResult {
  sendMessage: (prompt: string) => Promise<void>;
  stopStreaming: () => void;
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
  const store = useCaretStore();
  const settings = useCaretSettingsStore();
  const abortRef = useRef<AbortController | null>(null);

  const stopStreaming = useCallback((): void => {
    abortRef.current?.abort();
    abortRef.current = null;
    store.setStreaming(false);
  }, [store]);

  const sendMessage = useCallback(
    async (prompt: string): Promise<void> => {
      if (!prompt.trim() || store.isStreaming) return;

      const userMessageId = `user-${Date.now()}`;
      store.addMessage({
        id: userMessageId,
        role: 'user',
        blocks: [],
        content: prompt.trim(),
        createdAt: new Date().toISOString(),
      });

      store.startAssistantMessage();
      const assistantMessageId = `assistant-${Date.now()}`;

      const baseUrl = apiBaseUrl.endsWith('/') ? apiBaseUrl : `${apiBaseUrl}/`;
      const url = `${baseUrl}ai/generate-text-stream`;
      const controller = new AbortController();
      abortRef.current = controller;

      const openThinking: Record<string, OpenThinking> = {};

      try {
        const token = readAuthToken();
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
            session_id: store.sessionId ?? '',
            user_id: userId,
            system_context: systemContext ?? '',
            page_context: pageContext,
            mode: settings.mode,
          }),
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        for await (const event of consumeSse(response, assistantMessageId)) {
          dispatch(event, openThinking);
        }

        store.setStreaming(false);
      } catch (error) {
        if ((error as Error).name === 'AbortError') return;
        store.setStreaming(false);
        const message =
          error instanceof Error
            ? error.message
            : 'An unexpected error occurred. Please try again.';
        store.appendBlockToLastAssistant({
          id: `error-${Date.now()}`,
          type: 'text',
          content: `\n\n_Error: ${message}_`,
        });
      } finally {
        abortRef.current = null;
      }
    },
    [apiBaseUrl, userId, systemContext, pageContext, settings.mode, store]
  );

  /** Translate one server event into store mutations. */
  function dispatch(event: CaretStreamEvent, openThinking: Record<string, OpenThinking>): void {
    switch (event.type) {
      case 'session':
        if (!store.sessionId) store.setSessionId(event.sessionId);
        break;

      case 'text_delta':
        store.appendToLastAssistantMessage(event.delta);
        break;

      case 'thinking_start':
        openThinking[event.blockId] = { blockId: event.blockId, startedAt: Date.now() };
        store.appendBlockToLastAssistant({
          id: event.blockId,
          type: 'thinking',
          content: '',
        });
        break;

      case 'thinking_delta':
        store.updateBlockOnLastAssistant(event.blockId, (b) =>
          b.type === 'thinking' ? { ...b, content: b.content + event.delta } : b
        );
        break;

      case 'thinking_end': {
        const opened = openThinking[event.blockId];
        const duration = event.durationMs ?? (opened ? Date.now() - opened.startedAt : undefined);
        store.updateBlockOnLastAssistant(event.blockId, (b) =>
          b.type === 'thinking' ? { ...b, durationMs: duration } : b
        );
        delete openThinking[event.blockId];
        break;
      }

      case 'tool_call':
        store.appendBlockToLastAssistant({
          id: event.blockId,
          type: 'tool_call',
          name: event.name,
          input: event.input,
          status: event.requiresApproval ? 'pending' : 'running',
          preview: event.preview,
          requiresApproval: event.requiresApproval,
        });
        break;

      case 'tool_result':
        store.setToolStatus(event.blockId, event.status, {
          output: event.output,
          summary: event.summary,
          error: event.error,
        });
        break;

      case 'message_end':
      case 'done':
        store.setStreaming(false);
        break;

      case 'error':
        store.appendBlockToLastAssistant({
          id: `error-${Date.now()}`,
          type: 'text',
          content: `\n\n_${event.error}_`,
        });
        if (!event.recoverable) store.setStreaming(false);
        break;
    }
  }

  return { sendMessage, stopStreaming };
}
