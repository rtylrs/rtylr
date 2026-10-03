import { useCallback } from "react";
import type { CaretImageAttachment, CaretMessage } from "../types/message";
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
  return {
    id: `hist-${index}`,
    role,
    blocks: [{ id: `hist-${index}-text`, type: "text", content: m.content }],
    createdAt: m.createdAt ?? new Date().toISOString(),
    content: m.content,
    attachments,
  };
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
