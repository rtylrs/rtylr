import { describe, expect, it } from "vitest";
import { parseStreamLine } from "./events";

describe("parseStreamLine", () => {
  it("reads the server's snake_case fields", () => {
    expect(
      parseStreamLine('{"type":"session","session_id":"chat-1"}', "fallback"),
    ).toEqual({ type: "session", sessionId: "chat-1" });
    expect(
      parseStreamLine(
        '{"type":"tool_call","message_id":"m1","block_id":"tool-1","name":"memory.remember","requires_approval":true}',
        "fallback",
      ),
    ).toEqual({
      type: "tool_call",
      messageId: "m1",
      blockId: "tool-1",
      name: "memory.remember",
      requiresApproval: true,
    });
    expect(
      parseStreamLine(
        '{"type":"message_end","message_id":"m1","stored_id":"42"}',
        "fallback",
      ),
    ).toEqual({ type: "message_end", messageId: "m1", stored_id: "42" });
  });
});
