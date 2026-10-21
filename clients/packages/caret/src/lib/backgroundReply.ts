import type { CaretMessage } from "../types/message";

export const PENDING_REPLY_ID = "pending-background-reply";
export const BACKGROUND_REPLY_POLL_MS = 3000;
// Matches the server's limit on finishing a reply after its client leaves.
export const BACKGROUND_REPLY_WAIT_MS = 5 * 60 * 1000;

/** The chat ends with a recent message that has no reply yet. */
export function awaitsBackgroundReply(messages: CaretMessage[]): boolean {
  const last = messages[messages.length - 1];
  if (!last || last.role !== "user") return false;
  const sentAt = Date.parse(last.createdAt);
  return Number.isFinite(sentAt) && Date.now() - sentAt < BACKGROUND_REPLY_WAIT_MS;
}

export function withPendingReply(messages: CaretMessage[]): CaretMessage[] {
  if (!awaitsBackgroundReply(messages)) return messages;
  return [
    ...messages,
    {
      id: PENDING_REPLY_ID,
      role: "assistant",
      blocks: [{ id: `${PENDING_REPLY_ID}-text`, type: "text", content: "_Still answering…_" }],
      content: "",
      createdAt: new Date().toISOString(),
    },
  ];
}
