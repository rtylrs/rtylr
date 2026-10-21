import { describe, expect, it } from "vitest";
import { awaitsBackgroundReply, PENDING_REPLY_ID, withPendingReply } from "./backgroundReply";
import type { CaretMessage } from "../types/message";

const message = (role: "user" | "assistant", minutesAgo: number): CaretMessage => ({
  id: `${role}-${minutesAgo}`,
  role,
  blocks: [],
  content: "hi",
  createdAt: new Date(Date.now() - minutesAgo * 60_000).toISOString(),
});

describe("background replies", () => {
  it("waits for a reply to a recent unanswered message", () => {
    const loaded = [message("user", 0.5)];
    expect(awaitsBackgroundReply(loaded)).toBe(true);
    const shown = withPendingReply(loaded);
    expect(shown[shown.length - 1]?.id).toBe(PENDING_REPLY_ID);
  });

  it("does not wait when the chat is answered or the message is old", () => {
    expect(awaitsBackgroundReply([message("user", 1), message("assistant", 0.5)])).toBe(false);
    expect(awaitsBackgroundReply([message("user", 30)])).toBe(false);
    expect(withPendingReply([message("user", 30)])).toHaveLength(1);
  });
});
