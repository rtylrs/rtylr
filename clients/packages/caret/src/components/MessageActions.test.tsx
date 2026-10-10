import React from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import MessageActions, { formatResponseTime } from "./MessageActions";

describe("formatResponseTime", () => {
  it.each([
    [820, "0.8s"],
    [4200, "4.2s"],
    [15_400, "15s"],
    [65_000, "1m 05s"],
  ])("formats %ims as %s", (ms, label) => {
    expect(formatResponseTime(ms)).toBe(label);
  });
});

describe("MessageActions reply time", () => {
  it("shows the reply time beside copy and read aloud", () => {
    render(<MessageActions text="Hey!" responseMs={15_400} />);
    expect(screen.getByLabelText("Replied in 15s").textContent).toBe("15s");
  });

  it("shows nothing for replies that were not timed", () => {
    render(<MessageActions text="Hey!" />);
    expect(screen.queryByLabelText(/Replied in/)).toBeNull();
  });
});

describe("MessageActions rating", () => {
  it("rates up, clears on a second click, and asks why on a thumbs down", async () => {
    const { fireEvent } = await import("@testing-library/react");
    const calls: Array<[string, string | undefined]> = [];
    render(
      <MessageActions
        text="Hey!"
        onRate={(rating, reason) => calls.push([rating, reason])}
      />,
    );
    fireEvent.click(screen.getByLabelText("Good reply"));
    fireEvent.click(screen.getByLabelText("Good reply"));
    fireEvent.click(screen.getByLabelText("Bad reply"));
    const input = screen.getByLabelText("What was wrong with this reply?");
    fireEvent.change(input, { target: { value: "too long" } });
    fireEvent.keyDown(input, { key: "Enter" });
    fireEvent.blur(input);
    expect(calls).toEqual([
      ["up", undefined],
      ["clear", undefined],
      ["down", undefined],
      ["down", "too long"],
    ]);
  });

  it("hides rating buttons without a handler", () => {
    render(<MessageActions text="Hey!" />);
    expect(screen.queryByLabelText("Good reply")).toBeNull();
  });
});
