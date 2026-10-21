import { useCallback } from "react";
import type {
  CaretImageAttachment,
  CaretMessage,
  ContentBlock,
} from "../types/message";
import { createMemoryClient } from "../lib/memoryClient";
import type { ChatMessage } from "@voxire/voxie/hub/memory/v1/types";

interface UseLoadCaretChatParams {
  userId: number;
  apiBaseUrl: string;
}

function toCaretMessage(
  m: ChatMessage & { content: string },
  index: number,
): CaretMessage {
  const role = m.role === "assistant" || m.role === "system" ? m.role : "user";
  let attachments: CaretImageAttachment[] | undefined;
  try {
    const parsed: unknown = m.attachmentsJson
      ? JSON.parse(m.attachmentsJson)
      : [];
    if (Array.isArray(parsed))
      attachments = parsed
        .filter(isStoredImage)
        .map(({ name, data_url }) => ({ name, dataUrl: data_url }));
  } catch {
    attachments = undefined;
  }
  const stored = role === "assistant" ? storedBlocks(m.blocksJson, index) : [];
  return {
    id: `hist-${index}`,
    role,
    blocks:
      stored.length > 0
        ? stored
        : [{ id: `hist-${index}-text`, type: "text", content: m.content }],
    createdAt: m.createdAt ?? new Date().toISOString(),
    content: m.content,
    attachments,
    responseMs: m.responseMs && m.responseMs > 0 ? m.responseMs : undefined,
  };
}

const toolStatuses = new Set(["success", "error", "denied"]);

/**
 * Rebuild a saved reply's text and finished tool steps in the order they
 * streamed. Anything malformed is skipped; an empty result falls back to the
 * plain text.
 */
function storedBlocks(raw: string | undefined, index: number): ContentBlock[] {
  if (!raw) return [];
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return [];
  }
  if (!Array.isArray(parsed)) return [];

  const blocks: ContentBlock[] = [];
  parsed.forEach((value: unknown, position) => {
    if (!value || typeof value !== "object") return;
    const block = value as Record<string, unknown>;
    const id = `hist-${index}-${position}`;
    if (block.type === "text" && typeof block.content === "string") {
      if (block.content) blocks.push({ id, type: "text", content: block.content });
      return;
    }
    if (
      block.type === "tool_call" &&
      typeof block.name === "string" &&
      typeof block.status === "string" &&
      toolStatuses.has(block.status)
    ) {
      blocks.push({
        id,
        type: "tool_call",
        name: block.name,
        input:
          block.input && typeof block.input === "object"
            ? (block.input as Record<string, unknown>)
            : {},
        status: block.status as "success" | "error" | "denied",
        output: block.output,
        summary: typeof block.summary === "string" ? block.summary : undefined,
        error: typeof block.error === "string" ? block.error : undefined,
      });
    }
  });
  return blocks.some((block) => block.type === "text") ? blocks : [];
}

function isStoredImage(
  value: unknown,
): value is { name: string; data_url: string } {
  if (!value || typeof value !== "object") return false;
  const image = value as Record<string, unknown>;
  return (
    typeof image.name === "string" &&
    image.name.length <= 200 &&
    typeof image.data_url === "string" &&
    /^data:image\/(?:jpeg|png|webp);base64,[a-z0-9+/]+=*$/i.test(
      image.data_url,
    ) &&
    image.data_url.length <= 3 * 1024 * 1024
  );
}

/** Returns a loader that fetches a past session's messages from the server. */
export function useLoadCaretChat({
  userId: _userId,
  apiBaseUrl,
}: UseLoadCaretChatParams): (sessionId: string) => Promise<CaretMessage[]> {
  return useCallback(
    async (sessionId: string): Promise<CaretMessage[]> => {
      const res = await createMemoryClient(apiBaseUrl).getChatMessages({
        sessionId,
      });
      return (res.messages ?? [])
        .filter((m): m is ChatMessage & { content: string } =>
          Boolean(m.content),
        )
        .map(toCaretMessage);
    },
    [apiBaseUrl],
  );
}
