import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { CaretTab } from "../stores/caretStore";
import ChatTabs from "./ChatTabs";

const tabs: CaretTab[] = [
  {
    id: "revenue",
    sessionId: null,
    title: "Review revenue",
    projectId: null,
    messages: [],
    isStreaming: false,
    isLoading: false,
    isLoaded: true,
  },
  {
    id: "inventory",
    sessionId: null,
    title: "Check inventory",
    projectId: null,
    messages: [],
    isStreaming: true,
    isLoading: false,
    isLoaded: true,
  },
];

describe("ChatTabs", () => {
  it("exposes real tab semantics and supports arrow-key switching", () => {
    const onSelect = vi.fn();

    render(
      <ChatTabs
        tabs={tabs}
        activeTabId="revenue"
        onSelect={onSelect}
        onClose={vi.fn()}
      />,
    );

    const revenue = screen.getByRole("tab", { name: "Review revenue" });
    expect(revenue.getAttribute("aria-selected")).toBe("true");
    expect(revenue.getAttribute("aria-controls")).toBe(
      "caret-chat-panel-revenue",
    );
    expect(screen.getByText("Review revenue").getAttribute("title")).toBeNull();

    fireEvent.keyDown(revenue, { key: "ArrowRight" });
    expect(onSelect).toHaveBeenCalledWith("inventory");
    expect(document.activeElement).toBe(
      screen.getByRole("tab", { name: "Check inventory" }),
    );
  });

  it("supports closing the focused tab with Delete", () => {
    const onClose = vi.fn();

    render(
      <ChatTabs
        tabs={tabs}
        activeTabId="revenue"
        onSelect={vi.fn()}
        onClose={onClose}
      />,
    );

    fireEvent.keyDown(screen.getByRole("tab", { name: "Review revenue" }), {
      key: "Delete",
    });
    expect(onClose).toHaveBeenCalledWith("revenue");
  });
});
