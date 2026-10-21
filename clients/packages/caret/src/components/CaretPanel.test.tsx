import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import CaretPanel from "./CaretPanel";
import { useCaretStore } from "../stores/caretStore";
import { useCaretSettingsStore } from "../stores/settingsStore";
import { useCaretUIStore } from "../stores/uiStore";

const memoryClient = vi.hoisted(() => ({
  getChatMessages: vi.fn(),
  listChats: vi.fn(),
  listProjects: vi.fn(),
  deleteChat: vi.fn(),
}));

vi.mock("../lib/memoryClient", () => ({
  createMemoryClient: () => memoryClient,
}));

type HistoryResolver = (value: {
  messages: Array<{ role: string; content: string; createdAt: string }>;
}) => void;

let resolveHistory: HistoryResolver | null = null;

function renderPanel(): void {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  render(
    <QueryClientProvider client={queryClient}>
      <CaretPanel appId="dash" userId={1} apiBaseUrl="http://127.0.0.1:9" />
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  resolveHistory = null;
  useCaretUIStore.setState({ view: "chat" });
  useCaretSettingsStore.setState({ voiceEnabled: false, mode: "ask" });
  memoryClient.listChats.mockResolvedValue({ chats: [] });
  memoryClient.listProjects.mockResolvedValue({ projects: [] });
  memoryClient.getChatMessages.mockImplementation(
    () =>
      new Promise((resolve) => {
        resolveHistory = resolve;
      }),
  );

  const state = useCaretStore.getState();
  for (const tab of state.tabs.slice(1)) state.closeTab(tab.id);
  const emptyTabId = useCaretStore.getState().activeTabId;
  useCaretStore.getState().reset(emptyTabId);
  // Open a saved chat: its transcript is fetched when the panel mounts.
  useCaretStore
    .getState()
    .newTab({ sessionId: "chat-1", title: "Weekly report" });
  useCaretStore.getState().open();
});

afterEach(() => vi.unstubAllGlobals());

describe("CaretPanel new chats", () => {
  it.each([
    { view: "chat" as const, projectId: null },
    { view: "history" as const, projectId: "project-123" },
  ])(
    "creates a sendable chat from the $view button and preserves its project",
    async ({ view, projectId }) => {
      useCaretStore.getState().newTab({ projectId });
      useCaretUIStore.getState().setView(view);
      const fetchMock = vi
        .fn()
        .mockResolvedValue(
          new Response(
            'data: {"type":"session","sessionId":"new-session"}\n\ndata: {"type":"text_delta","messageId":"reply","delta":"Hello there."}\n\ndata: {"type":"done"}\n\n',
            { headers: { "Content-Type": "text/event-stream" } },
          ),
        );
      vi.stubGlobal("fetch", fetchMock);
      renderPanel();

      fireEvent.click(
        view === "chat"
          ? screen.getByLabelText("New chat")
          : screen.getByRole("button", { name: "New chat" }),
      );
      fireEvent.change(screen.getByLabelText("Message Caret"), {
        target: { value: "Hello" },
      });
      fireEvent.click(screen.getByRole("button", { name: "Send message" }));

      await waitFor(() =>
        expect(screen.getByText("Hello there.")).not.toBeNull(),
      );
      expect(fetchMock).toHaveBeenCalledTimes(1);
      const request = JSON.parse(fetchMock.mock.calls[0][1].body);
      expect(request.project_id).toBe(projectId ?? "");
      const persisted = JSON.parse(
        localStorage.getItem("caret-session") ?? "{}",
      );
      expect(
        persisted.state.tabs.find(
          (tab: { sessionId: string | null }) =>
            tab.sessionId === "new-session",
        ).projectId,
      ).toBe(projectId);
    },
  );
});

describe("CaretPanel history loading", () => {
  it("keeps a closed panel out of keyboard navigation and focuses the composer when opened", () => {
    useCaretStore.getState().close();
    render(<button>Page action</button>);
    const button = screen.getByRole("button", { name: "Page action" });
    button.focus();
    renderPanel();
    expect(document.activeElement).toBe(button);
    expect(
      screen.queryByRole("complementary", { name: "Caret AI assistant" }),
    ).toBeNull();
    act(() => useCaretStore.getState().open());
    expect(document.activeElement).toBe(
      screen.getByRole("textbox", { name: "Message Caret" }),
    );
  });
  it("keeps a prompt sent while the history request is in flight", async () => {
    renderPanel();

    const tabId = useCaretStore.getState().activeTabId;
    await waitFor(() =>
      expect(memoryClient.getChatMessages).toHaveBeenCalledTimes(1),
    );

    fireEvent.change(screen.getByLabelText("Message Caret"), {
      target: { value: "What changed this week?" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Send message" }));
    await waitFor(() => {
      const tab = useCaretStore
        .getState()
        .tabs.find((candidate) => candidate.id === tabId);
      expect(
        tab?.messages.some(
          (message) => message.content === "What changed this week?",
        ),
      ).toBe(true);
    });

    await act(async () => {
      resolveHistory?.({
        messages: [
          {
            role: "user",
            content: "Previous question about revenue",
            createdAt: "2026-09-18T10:00:00.000Z",
          },
          {
            role: "assistant",
            content: "Previous answer about revenue",
            createdAt: "2026-09-18T10:00:05.000Z",
          },
        ],
      });
    });

    const tab = useCaretStore
      .getState()
      .tabs.find((candidate) => candidate.id === tabId);
    expect(
      tab?.messages.some(
        (message) => message.content === "What changed this week?",
      ),
    ).toBe(true);
    expect(
      tab?.messages.some(
        (message) => message.content === "Previous answer about revenue",
      ),
    ).toBe(false);
    expect(tab?.isLoaded).toBe(true);
    // Marking the tab loaded must stop the effect from refetching the history.
    expect(memoryClient.getChatMessages).toHaveBeenCalledTimes(1);
  });

  it("surfaces a failed history request and retries it from the history view", async () => {
    memoryClient.listChats.mockReset();
    memoryClient.listChats.mockRejectedValueOnce(
      new Error("history service unavailable"),
    );
    memoryClient.listChats.mockResolvedValueOnce({
      chats: [
        {
          id: "chat-2",
          title: "Recovered chat",
          updatedAt: "2026-09-20T10:00:00.000Z",
        },
      ],
    });
    memoryClient.getChatMessages.mockResolvedValue({ messages: [] });
    useCaretUIStore.getState().setView("history");

    renderPanel();

    await waitFor(() =>
      expect(screen.getByRole("alert").textContent).toContain(
        "Unable to load saved chats.",
      ),
    );
    expect(screen.queryByText("No saved chats yet.")).toBeNull();

    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    await waitFor(() =>
      expect(screen.getByText("Recovered chat")).not.toBeNull(),
    );
  });
});
