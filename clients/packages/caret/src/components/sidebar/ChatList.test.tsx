import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import ChatList from "./ChatList";

describe("ChatList history loading", () => {
  it("shows a retryable load error instead of claiming history is empty", () => {
    const onRetry = vi.fn();

    render(
      <ChatList
        chats={[]}
        activeId={null}
        onSelect={vi.fn()}
        onNew={vi.fn()}
        error={new Error("history service unavailable")}
        onRetry={onRetry}
      />,
    );

    expect(screen.getByRole("alert").textContent).toContain(
      "Unable to load saved chats.",
    );
    expect(screen.queryByText("No saved chats yet.")).toBeNull();

    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(onRetry).toHaveBeenCalledOnce();
  });
});

describe("ChatList delete confirmation", () => {
  const chat = {
    id: "chat-1",
    title: "Saudi Arabia visa",
    pinned: false,
    updatedAt: "2026-10-12T10:00:00Z",
  } as never;

  const renderList = (onDelete: () => Promise<void>) =>
    render(
      <ChatList
        chats={[chat]}
        activeId={null}
        onSelect={vi.fn()}
        onNew={vi.fn()}
        onDelete={onDelete}
      />,
    );

  it("asks before deleting and keeps the chat when cancelled", () => {
    const onDelete = vi.fn().mockResolvedValue(undefined);
    renderList(onDelete);

    fireEvent.click(screen.getByRole("button", { name: "Delete chat" }));
    expect(screen.getByRole("alertdialog").textContent).toContain(
      "Saudi Arabia visa",
    );
    expect(onDelete).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("alertdialog")).toBeNull();
    expect(onDelete).not.toHaveBeenCalled();
  });

  it("deletes only after confirming", async () => {
    const onDelete = vi.fn().mockResolvedValue(undefined);
    renderList(onDelete);

    fireEvent.click(screen.getByRole("button", { name: "Delete chat" }));
    fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(onDelete).toHaveBeenCalledWith("chat-1"));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
  });
});
