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

const chat = {
  id: "chat-1",
  title: "Saudi Arabia visa",
  pinned: false,
  updatedAt: "2026-10-12T10:00:00Z",
} as never;

function openActions(): void {
  fireEvent.click(screen.getByRole("button", { name: "Chat actions" }));
}

describe("ChatList delete confirmation", () => {
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

    openActions();
    fireEvent.click(screen.getByRole("menuitem", { name: "Delete chat" }));
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

    openActions();
    fireEvent.click(screen.getByRole("menuitem", { name: "Delete chat" }));
    fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(onDelete).toHaveBeenCalledWith("chat-1"));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
  });
});

describe("ChatList chat actions", () => {
  const renderList = (props: Partial<React.ComponentProps<typeof ChatList>>) =>
    render(
      <ChatList
        chats={[chat]}
        activeId={null}
        onSelect={vi.fn()}
        onNew={vi.fn()}
        {...props}
      />,
    );

  it("opens the rename dialog on double click and saves the new title", async () => {
    const onRename = vi.fn().mockResolvedValue(undefined);
    renderList({ onRename });

    fireEvent.doubleClick(screen.getByText("Saudi Arabia visa"));
    const input = screen.getByRole("textbox", { name: "Rename chat" });
    expect((input as HTMLInputElement).value).toBe("Saudi Arabia visa");

    fireEvent.change(input, { target: { value: "  Visa checklist " } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() =>
      expect(onRename).toHaveBeenCalledWith("chat-1", "Visa checklist"),
    );
    await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
  });

  it("keeps the dialog open with an error when the rename fails", async () => {
    const onRename = vi.fn().mockRejectedValue(new Error("boom"));
    renderList({ onRename });

    openActions();
    fireEvent.click(screen.getByRole("menuitem", { name: "Rename" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Rename chat" }), {
      target: { value: "New name" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect((await screen.findByRole("alert")).textContent).toContain(
      "Unable to rename this chat.",
    );
    expect(screen.getByRole("dialog")).not.toBeNull();
  });

  it("doesn't save an empty name and closes on Escape without renaming", () => {
    const onRename = vi.fn();
    renderList({ onRename });

    fireEvent.doubleClick(screen.getByText("Saudi Arabia visa"));
    fireEvent.change(screen.getByRole("textbox", { name: "Rename chat" }), {
      target: { value: "   " },
    });
    expect(
      (screen.getByRole("button", { name: "Save" }) as HTMLButtonElement)
        .disabled,
    ).toBe(true);

    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(onRename).not.toHaveBeenCalled();
  });

  it("pins from the menu and offers Unpin for a pinned chat", () => {
    const onTogglePin = vi.fn();
    const { unmount } = renderList({ onTogglePin });
    openActions();
    fireEvent.click(screen.getByRole("menuitem", { name: "Pin chat" }));
    expect(onTogglePin).toHaveBeenCalledWith("chat-1");
    unmount();

    renderList({ chats: [{ ...(chat as object), pinned: true } as never], onTogglePin });
    openActions();
    expect(screen.getByRole("menuitem", { name: "Unpin chat" })).not.toBeNull();
  });

  it("opens the same menu with a right click", () => {
    renderList({ onRename: vi.fn() });
    fireEvent.contextMenu(screen.getByText("Saudi Arabia visa"));
    expect(screen.getByRole("menuitem", { name: "Rename" })).not.toBeNull();
  });
});
