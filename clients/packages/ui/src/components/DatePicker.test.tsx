import React from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { DatePicker } from "../index";
import { placePopup } from "./DatePicker";

describe("DatePicker", () => {
  it("shows a placeholder when there is no value", () => {
    render(
      <DatePicker onChange={vi.fn()} labels={{ placeholder: "Select date" }} />,
    );
    expect(screen.getByRole("button")).toHaveTextContent("Select date");
  });

  it("shows the formatted value when one is set", () => {
    render(<DatePicker value="2026-08-16" onChange={vi.fn()} locale="en-US" />);
    expect(screen.getByRole("button")).toHaveTextContent("Aug 16, 2026");
  });

  it("opens a calendar panel on click and closes it after picking a date", () => {
    const handleChange = vi.fn();
    render(
      <DatePicker value="2026-08-16" onChange={handleChange} locale="en-US" />,
    );

    fireEvent.click(screen.getByRole("button", { name: /aug 16, 2026/i }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("gridcell", { name: /august 20, 2026/i }));
    expect(handleChange).toHaveBeenCalledWith("2026-08-20");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("closes the panel on Escape without changing the value", () => {
    const handleChange = vi.fn();
    render(
      <DatePicker value="2026-08-16" onChange={handleChange} locale="en-US" />,
    );
    fireEvent.click(screen.getByRole("button"));
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    fireEvent.keyDown(document, { key: "Escape" });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(handleChange).not.toHaveBeenCalled();
  });

  it("closes the panel on an outside click", () => {
    render(
      <div>
        <DatePicker value="2026-08-16" onChange={vi.fn()} locale="en-US" />
        <button>outside</button>
      </div>,
    );
    fireEvent.click(screen.getByRole("button", { name: /aug 16, 2026/i }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    fireEvent.mouseDown(screen.getByRole("button", { name: "outside" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("shows a clear control only when clearable and a value is set, and it does not open the panel", () => {
    const handleChange = vi.fn();
    render(
      <DatePicker
        value="2026-08-16"
        onChange={handleChange}
        locale="en-US"
        clearable
        labels={{ clear: "Clear date" }}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Clear date" }));
    expect(handleChange).toHaveBeenCalledWith("");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("does not open when disabled", () => {
    render(
      <DatePicker
        value="2026-08-16"
        onChange={vi.fn()}
        locale="en-US"
        disabled
      />,
    );
    fireEvent.click(screen.getByRole("button"));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("is reachable via a label using its id, like a native form control", () => {
    render(
      <div>
        <label htmlFor="due-date">Due date</label>
        <DatePicker
          id="due-date"
          value="2026-08-16"
          onChange={vi.fn()}
          locale="en-US"
        />
      </div>,
    );
    expect(screen.getByLabelText("Due date")).toHaveAttribute(
      "data-value",
      "2026-08-16",
    );
  });

  // The panel hung below the field inside its parent, so a scrolling dialog
  // with a footer cut its last week off. It now opens in the top layer.
  it("opens the panel as a top-layer popover so a scrolling parent cannot clip it", () => {
    render(<DatePicker value="2026-08-16" onChange={vi.fn()} locale="en-US" />);
    fireEvent.click(screen.getByRole("button", { name: /aug 16, 2026/i }));
    expect(screen.getByRole("dialog")).toHaveAttribute("popover", "manual");
  });
});

describe("placePopup", () => {
  const viewport = { width: 1400, height: 720 };
  const size = { width: 280, height: 300 };

  it("opens below the field when there is room", () => {
    const anchor = { top: 100, bottom: 140, left: 200, right: 400 };
    expect(placePopup(anchor, size, viewport, "ltr")).toEqual({
      top: 144,
      left: 200,
    });
  });

  it("opens above the field when the room below would cut it off", () => {
    const anchor = { top: 560, bottom: 600, left: 200, right: 400 };
    expect(placePopup(anchor, size, viewport, "ltr")).toEqual({
      top: 256,
      left: 200,
    });
  });

  it("aligns to the field end in right-to-left pages", () => {
    const anchor = { top: 100, bottom: 140, left: 1100, right: 1300 };
    expect(placePopup(anchor, size, viewport, "rtl").left).toBe(1020);
  });

  it("stays on screen near the right edge", () => {
    const anchor = { top: 100, bottom: 140, left: 1200, right: 1390 };
    expect(placePopup(anchor, size, viewport, "ltr").left).toBe(1108);
  });
});
