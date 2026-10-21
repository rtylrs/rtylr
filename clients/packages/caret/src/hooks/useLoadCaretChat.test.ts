import { describe, expect, it, vi } from "vitest";
import { renderHook } from "@testing-library/react";

const getChatMessages = vi.fn();
vi.mock("../lib/memoryClient", () => ({
  createMemoryClient: () => ({ getChatMessages }),
}));

import { useLoadCaretChat } from "./useLoadCaretChat";

describe("useLoadCaretChat", () => {
  it("restores a saved reply's search step, text order, and reply time", async () => {
    getChatMessages.mockResolvedValue({
      messages: [
        { role: "user", content: "news?" },
        {
          role: "assistant",
          content: "I'll search.\nHere's what I found.",
          responseMs: 7600,
          blocksJson: JSON.stringify([
            { type: "text", content: "I'll search." },
            {
              type: "tool_call",
              name: "web.search",
              status: "success",
              summary: "3 results",
              input: { query: "news" },
            },
            { type: "tool_call", name: "crm.update", status: "pending" },
            { type: "text", content: "Here's what I found." },
          ]),
        },
      ],
    });

    const { result } = renderHook(() =>
      useLoadCaretChat({ userId: 1, apiBaseUrl: "" }),
    );
    const [user, assistant] = await result.current("session-1");

    expect(user?.blocks).toHaveLength(1);
    expect(assistant?.responseMs).toBe(7600);
    expect(assistant?.blocks.map((block) => block.type)).toEqual([
      "text",
      "tool_call",
      "text",
    ]);
    expect(assistant?.blocks[1]).toMatchObject({
      name: "web.search",
      status: "success",
      summary: "3 results",
    });
  });

  it("falls back to plain text for replies saved before steps were stored", async () => {
    getChatMessages.mockResolvedValue({
      messages: [{ role: "assistant", content: "Old reply", blocksJson: "" }],
    });

    const { result } = renderHook(() =>
      useLoadCaretChat({ userId: 1, apiBaseUrl: "" }),
    );
    const [assistant] = await result.current("session-1");

    expect(assistant?.blocks).toEqual([
      { id: "hist-0-text", type: "text", content: "Old reply" },
    ]);
    expect(assistant?.responseMs).toBeUndefined();
  });
});
