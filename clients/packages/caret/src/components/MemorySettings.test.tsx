import { describe, expect, it, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";

const save = vi.fn();
const remove = vi.fn();
let memories = [{ id: "1", value: "Prefers short answers" }];

vi.mock("../hooks/useCaretMemories", () => ({
  useCaretMemories: () => ({
    memories,
    isLoading: false,
    error: null,
    save,
    remove,
    isSaving: false,
  }),
}));

import MemorySettings from "./MemorySettings";

beforeEach(() => {
  save.mockReset().mockResolvedValue({});
  remove.mockReset().mockResolvedValue({});
  memories = [{ id: "1", value: "Prefers short answers" }];
});

describe("MemorySettings", () => {
  it("adds a memory", async () => {
    render(<MemorySettings userId={1} apiBaseUrl="" />);
    const input = screen.getByLabelText("Add something for Caret to remember");
    fireEvent.change(input, {
      target: { value: "  Our fiscal year starts in April " },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add memory" }));

    await waitFor(() =>
      expect(save).toHaveBeenCalledWith("Our fiscal year starts in April"),
    );
    await waitFor(() => expect((input as HTMLInputElement).value).toBe(""));
  });

  it("edits and forgets a memory", async () => {
    render(<MemorySettings userId={1} apiBaseUrl="" />);

    fireEvent.click(
      screen.getByRole("button", {
        name: "Edit memory: Prefers short answers",
      }),
    );
    const editor = screen.getByLabelText("Edit memory");
    fireEvent.change(editor, { target: { value: "Prefers detailed answers" } });
    fireEvent.keyDown(editor, { key: "Enter" });
    await waitFor(() =>
      expect(save).toHaveBeenCalledWith("Prefers detailed answers", "1"),
    );

    fireEvent.click(
      screen.getByRole("button", { name: "Forget: Prefers short answers" }),
    );
    await waitFor(() => expect(remove).toHaveBeenCalledWith("1"));
  });

  it("explains how to add memories when empty and reports failures", async () => {
    memories = [];
    save.mockRejectedValue(new Error("nope"));
    render(<MemorySettings userId={1} apiBaseUrl="" />);
    expect(screen.getByText(/remember that/)).toBeTruthy();

    fireEvent.change(
      screen.getByLabelText("Add something for Caret to remember"),
      {
        target: { value: "x" },
      },
    );
    fireEvent.click(screen.getByRole("button", { name: "Add memory" }));
    expect(await screen.findByRole("alert")).toBeTruthy();
  });
});
