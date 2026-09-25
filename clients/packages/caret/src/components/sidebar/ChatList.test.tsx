import { fireEvent, render, screen } from "@testing-library/react";
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
