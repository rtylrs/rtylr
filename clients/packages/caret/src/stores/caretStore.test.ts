import { beforeEach, describe, expect, it } from "vitest";
import { useCaretStore } from "./caretStore";
import type { CaretMessage } from "../types/message";

function userMessage(content: string): CaretMessage {
  return {
    id: `user-${content}`,
    role: "user",
    blocks: [],
    content,
    createdAt: new Date().toISOString(),
  };
}

beforeEach(() => {
  const state = useCaretStore.getState();
  for (const tab of state.tabs.slice(1)) state.closeTab(tab.id);
  const activeTabId = useCaretStore.getState().activeTabId;
  useCaretStore.getState().reset(activeTabId);
  useCaretStore.getState().setActiveTab(activeTabId);
});

describe("useCaretStore tabs", () => {
  it("keeps messages and streams isolated per tab", () => {
    const firstTabId = useCaretStore.getState().activeTabId;
    useCaretStore.getState().addMessage(userMessage("first chat"), firstTabId);

    const secondTabId = useCaretStore.getState().newTab();
    useCaretStore
      .getState()
      .addMessage(userMessage("second chat"), secondTabId);
    useCaretStore.getState().setStreaming(true, secondTabId);

    const state = useCaretStore.getState();
    const firstTab = state.tabs.find((tab) => tab.id === firstTabId);
    const secondTab = state.tabs.find((tab) => tab.id === secondTabId);

    expect(firstTab?.messages.map((message) => message.content)).toEqual([
      "first chat",
    ]);
    expect(firstTab?.isStreaming).toBe(false);
    expect(secondTab?.messages.map((message) => message.content)).toEqual([
      "second chat",
    ]);
    expect(secondTab?.isStreaming).toBe(true);
  });

  it("derives a readable tab title from the first user message", () => {
    const tabId = useCaretStore.getState().activeTabId;
    useCaretStore
      .getState()
      .addMessage(userMessage("Review overdue invoices"), tabId);

    expect(useCaretStore.getState().tabs[0].title).toBe(
      "Review overdue invoices",
    );
  });

  it("persists the selected project with its chat tab", () => {
    const tabId = useCaretStore.getState().newTab({ projectId: "project-123" });
    const persisted = JSON.parse(
      localStorage.getItem("caret-session") ?? "{}",
    ) as {
      state?: { tabs?: Array<{ id: string; projectId?: string | null }> };
    };

    expect(
      persisted.state?.tabs?.find((tab) => tab.id === tabId)?.projectId,
    ).toBe("project-123");
  });
});

describe("useCaretStore reply timing", () => {
  it("records the reply time on the latest assistant message only", () => {
    const tabId = useCaretStore.getState().activeTabId;
    useCaretStore.getState().addMessage(userMessage("hi"), tabId);
    useCaretStore.getState().startAssistantMessage(tabId);
    useCaretStore.getState().setLastAssistantResponseTime(4200, tabId);

    const messages = useCaretStore
      .getState()
      .tabs.find((tab) => tab.id === tabId)!.messages;
    expect(messages[messages.length - 1]?.responseMs).toBe(4200);
    expect(messages[0].responseMs).toBeUndefined();
  });

  it("ignores a reply time when the last message is not from the assistant", () => {
    const tabId = useCaretStore.getState().activeTabId;
    useCaretStore.getState().addMessage(userMessage("hi"), tabId);
    useCaretStore.getState().setLastAssistantResponseTime(4200, tabId);

    const messages = useCaretStore
      .getState()
      .tabs.find((tab) => tab.id === tabId)!.messages;
    expect(messages[messages.length - 1]?.responseMs).toBeUndefined();
  });
});

describe("useCaretStore streamed text", () => {
  it("starts a new text block after a tool call so the reply keeps its order", () => {
    const tabId = useCaretStore.getState().activeTabId;
    const store = useCaretStore.getState();
    store.startAssistantMessage(tabId);
    store.appendToLastAssistantMessage("I'll search. ", tabId);
    store.appendToLastAssistantMessage("One moment.", tabId);
    store.appendBlockToLastAssistant(
      {
        id: "tool-1",
        type: "tool_call",
        name: "web.search",
        input: {},
        status: "running",
      },
      tabId,
    );
    store.appendToLastAssistantMessage("Here's what I found.", tabId);

    const messages = useCaretStore
      .getState()
      .tabs.find((tab) => tab.id === tabId)!.messages;
    const blocks = messages[messages.length - 1]!.blocks;
    expect(blocks.map((block) => block.type)).toEqual([
      "text",
      "tool_call",
      "text",
    ]);
    expect(blocks[0]).toMatchObject({ content: "I'll search. One moment." });
    expect(blocks[2]).toMatchObject({ content: "Here's what I found." });
    expect(new Set(blocks.map((block) => block.id)).size).toBe(3);
  });
});
