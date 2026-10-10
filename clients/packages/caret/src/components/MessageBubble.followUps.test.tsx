import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import MessageBubble from "./MessageBubble";
import type { CaretMessage } from "../types/message";

const reply: CaretMessage = {
  id: "a1",
  role: "assistant",
  blocks: [{ id: "t", type: "text", content: "Revenue grew 12%." }],
  content: "Revenue grew 12%.",
  createdAt: new Date().toISOString(),
  followUps: ["Which products drove it?", "Compare with last year"],
};

describe("MessageBubble follow-ups", () => {
  it("offers the suggested questions under the latest reply", () => {
    const onFollowUp = vi.fn();
    render(<MessageBubble message={reply} onFollowUp={onFollowUp} />);
    fireEvent.click(screen.getByRole("button", { name: "Which products drove it?" }));
    expect(onFollowUp).toHaveBeenCalledWith("Which products drove it?");
  });

  it("hides them on older replies and while streaming", () => {
    const { rerender } = render(<MessageBubble message={reply} />);
    expect(screen.queryByRole("button", { name: "Compare with last year" })).toBeNull();
    rerender(<MessageBubble message={reply} isStreaming onFollowUp={() => undefined} />);
    expect(screen.queryByRole("button", { name: "Compare with last year" })).toBeNull();
  });
});
