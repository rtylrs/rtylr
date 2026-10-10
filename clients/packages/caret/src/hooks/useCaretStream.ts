import { useCallback } from "react";
import { useCaretStore } from "../stores/caretStore";
import { consumeSse } from "../protocol/sse";
import type { CaretStreamEvent } from "../protocol/events";
import { readAuthToken } from "../lib/auth";
import {
  MAX_ATTACHMENT_PAYLOAD_BYTES,
  readTextFiles,
} from "../lib/fileAttachments";
import { prepareCaretImages } from "../lib/imageAttachments";
import type { CaretPageContext } from "../types/context";
import type { CaretMode } from "../types/mode";

export interface UseCaretStreamParams {
  userId: number;
  apiBaseUrl: string;
  systemContext?: string;
  pageContext?: CaretPageContext;
}

export interface UseCaretStreamResult {
  sendMessage: (
    prompt: string,
    modeOverride?: CaretMode,
    images?: File[],
  ) => Promise<void>;
  /**
   * Stops a reply. By default the server stops generating too; pass
   * { keepServer: true } to only stop listening (closing a chat tab), so the
   * reply still finishes and is saved to the chat.
   */
  stopStreaming: (tabId?: string, options?: { keepServer?: boolean }) => void;
}

interface OpenThinking {
  blockId: string;
  startedAt: number;
}

/**
 * A stream outlives the hook instance that started it when the panel unmounts
 * mid-response (navigating between layouts/routes), so the controllers are
 * shared at module scope. A remounted instance can then still stop the stream.
 */
const streamControllers = new Map<string, AbortController>();

export function useCaretStream({
  userId,
  apiBaseUrl,
  systemContext,
  pageContext,
}: UseCaretStreamParams): UseCaretStreamResult {
  const stopStreaming = useCallback(
    (tabId?: string, options?: { keepServer?: boolean }): void => {
    const store = useCaretStore.getState();
    const targetTabId = tabId ?? store.activeTabId;
    const controller = streamControllers.get(targetTabId);
    const sessionId = store.tabs.find((tab) => tab.id === targetTabId)?.sessionId;
    // A dropped connection keeps the reply generating on the server so it is
    // not lost; an explicit stop has to say so.
    if (controller && sessionId && !options?.keepServer) {
      const baseUrl = apiBaseUrl.endsWith("/") ? apiBaseUrl : `${apiBaseUrl}/`;
      const token = readAuthToken();
      void fetch(`${baseUrl}ai/generate-text-stream/stop`, {
        method: "POST",
        credentials: "include",
        keepalive: true,
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify({ session_id: sessionId }),
      }).catch(() => undefined);
    }
    controller?.abort();
    if (streamControllers.get(targetTabId) === controller) {
      streamControllers.delete(targetTabId);
    }
    store.setStreaming(false, targetTabId);
    },
    [apiBaseUrl],
  );

  const sendMessage = useCallback(
    async (
      prompt: string,
      modeOverride?: CaretMode,
      files: File[] = [],
    ): Promise<void> => {
      const store = useCaretStore.getState();
      const tab = store.tabs.find(
        (candidate) => candidate.id === store.activeTabId,
      );
      if ((!prompt.trim() && files.length === 0) || !tab || tab.isStreaming)
        return;

      const tabId = tab.id;

      const assistantMessageId = `assistant-${Date.now()}`;

      const baseUrl = apiBaseUrl.endsWith("/") ? apiBaseUrl : `${apiBaseUrl}/`;
      const url = `${baseUrl}ai/generate-text-stream`;
      const controller = new AbortController();
      streamControllers.set(tabId, controller);

      const openThinking: Record<string, OpenThinking> = {};
      let assistantStarted = false;
      // Anything typed as an image goes through image preparation, which
      // rejects formats Caret can't read.
      const imageFiles = files.filter((file) => file.type.startsWith("image/"));
      const textFiles = files.filter((file) => !file.type.startsWith("image/"));
      const requestPrompt =
        prompt.trim() ||
        (textFiles.length > 0
          ? "Please look at the attached file."
          : "Please describe the attached image.");

      try {
        const images = await prepareCaretImages(imageFiles);
        const texts = await readTextFiles(textFiles);
        const payloadBytes =
          images.reduce((total, image) => total + image.dataUrl.length, 0) +
          texts.reduce(
            (total, file) =>
              total + file.text.length * 2 + (file.data_url?.length ?? 0),
            0,
          );
        if (payloadBytes > MAX_ATTACHMENT_PAYLOAD_BYTES)
          throw new Error(
            "These attachments are too large to send together; attach fewer or smaller files.",
          );
        store.addMessage(
          {
            id: `user-${Date.now()}`,
            role: "user",
            blocks: [],
            content: requestPrompt,
            attachments: [
              ...images,
              ...texts.map(({ name }) => ({
                name,
                dataUrl: "",
                kind: "file" as const,
              })),
            ],
            createdAt: new Date().toISOString(),
          },
          tabId,
        );
        store.startAssistantMessage(tabId);
        assistantStarted = true;
        const sentAt = performance.now();
        const token = readAuthToken();
        const mode = modeOverride ?? "agent";
        const response = await fetch(url, {
          method: "POST",
          credentials: "include",
          signal: controller.signal,
          headers: {
            "Content-Type": "application/json",
            ...(token ? { Authorization: `Bearer ${token}` } : {}),
          },
          body: JSON.stringify({
            prompt: requestPrompt,
            session_id:
              useCaretStore
                .getState()
                .tabs.find((candidate) => candidate.id === tabId)?.sessionId ??
              "",
            project_id:
              useCaretStore
                .getState()
                .tabs.find((candidate) => candidate.id === tabId)?.projectId ??
              "",
            user_id: userId,
            system_context: systemContext ?? "",
            page_context: pageContext,
            mode,
            images: images.map(({ name, dataUrl }) => ({
              name,
              data_url: dataUrl,
            })),
            files: texts,
            timezone: browserTimezone(),
          }),
        });

        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        // Time the reply when it finishes, not when the connection closes: the
        // actions row appears on "done", and the server may hold the stream open.
        let timed = false;
        const recordReplyTime = (): void => {
          if (timed) return;
          timed = true;
          useCaretStore
            .getState()
            .setLastAssistantResponseTime(
              Math.round(performance.now() - sentAt),
              tabId,
            );
        };

        for await (const event of consumeSse(response, assistantMessageId)) {
          dispatch(event, openThinking, tabId);
          if (event.type === "done" || event.type === "message_end") {
            recordReplyTime();
          }
        }

        recordReplyTime();
        useCaretStore.getState().setStreaming(false, tabId);
      } catch (error) {
        if ((error as Error).name === "AbortError") return;
        useCaretStore.getState().setStreaming(false, tabId);
        const message =
          error instanceof Error
            ? error.message
            : "An unexpected error occurred. Please try again.";
        if (assistantStarted) {
          useCaretStore.getState().appendBlockToLastAssistant(
            {
              id: `error-${Date.now()}`,
              type: "text",
              content: `\n\n_Error: ${message}_`,
            },
            tabId,
          );
        } else {
          useCaretStore.getState().addMessage(
            {
              id: `assistant-image-error-${Date.now()}`,
              role: "assistant",
              blocks: [
                {
                  id: `image-error-${Date.now()}`,
                  type: "text",
                  content: message,
                },
              ],
              content: message,
              createdAt: new Date().toISOString(),
            },
            tabId,
          );
        }
      } finally {
        if (streamControllers.get(tabId) === controller) {
          streamControllers.delete(tabId);
          useCaretStore.getState().setStreaming(false, tabId);
        }
      }
    },
    [apiBaseUrl, userId, systemContext, pageContext],
  );

  /** Translate one server event into store mutations. */
  function dispatch(
    event: CaretStreamEvent,
    openThinking: Record<string, OpenThinking>,
    tabId: string,
  ): void {
    const store = useCaretStore.getState();
    switch (event.type) {
      case "session":
        if (
          !store.tabs.find((candidate) => candidate.id === tabId)?.sessionId
        ) {
          store.setSessionId(event.sessionId, tabId);
        }
        break;

      case "text_delta":
        store.appendToLastAssistantMessage(event.delta, tabId);
        break;

      case "thinking_start":
        openThinking[event.blockId] = {
          blockId: event.blockId,
          startedAt: Date.now(),
        };
        store.appendBlockToLastAssistant(
          {
            id: event.blockId,
            type: "thinking",
            content: "",
          },
          tabId,
        );
        break;

      case "thinking_delta":
        store.updateBlockOnLastAssistant(
          event.blockId,
          (b) =>
            b.type === "thinking"
              ? { ...b, content: b.content + event.delta }
              : b,
          tabId,
        );
        break;

      case "thinking_end": {
        const opened = openThinking[event.blockId];
        const duration =
          event.durationMs ??
          (opened ? Date.now() - opened.startedAt : undefined);
        store.updateBlockOnLastAssistant(
          event.blockId,
          (b) => (b.type === "thinking" ? { ...b, durationMs: duration } : b),
          tabId,
        );
        delete openThinking[event.blockId];
        break;
      }

      case "tool_call":
        store.appendBlockToLastAssistant(
          {
            id: event.blockId,
            type: "tool_call",
            name: event.name,
            mode: event.mode,
            input: event.input,
            status: event.requiresApproval ? "pending" : "running",
            preview: event.preview,
            requiresApproval: event.requiresApproval,
          },
          tabId,
        );
        break;

      case "tool_result":
        store.setToolStatus(
          event.blockId,
          event.status,
          {
            output: event.output,
            summary: event.summary,
            error: event.error,
          },
          tabId,
        );
        break;

      case "follow_ups":
        if (Array.isArray(event.items) && event.items.length > 0) {
          store.setLastAssistantFollowUps(event.items.slice(0, 3), tabId);
        }
        break;

      case "message_end":
        if (event.stored_id) store.setLastAssistantStoredId(event.stored_id, tabId);
        store.setStreaming(false, tabId);
        break;
      case "done":
        store.setStreaming(false, tabId);
        break;

      case "error":
        store.appendBlockToLastAssistant(
          {
            id: `error-${Date.now()}`,
            type: "text",
            content: `\n\n_${event.error}_`,
          },
          tabId,
        );
        if (!event.recoverable) store.setStreaming(false, tabId);
        break;
    }
  }

  return { sendMessage, stopStreaming };
}

/** The person's IANA timezone, so Caret's "today" is their today. */
function browserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone ?? "";
  } catch {
    return "";
  }
}
