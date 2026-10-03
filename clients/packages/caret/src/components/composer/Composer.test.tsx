import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Composer } from "./Composer";

describe("Composer image attachments", () => {
  it("accepts a dropped image, previews it, and sends it with the prompt", () => {
    const file = new File(["image bytes"], "receipt.png", {
      type: "image/png",
    });
    const onSubmit = vi.fn();
    const { container } = render(
      <Composer isStreaming={false} onSubmit={onSubmit} onStop={vi.fn()} />,
    );
    const composer = container.querySelector(".caret-composer");
    expect(composer).not.toBeNull();

    fireEvent.drop(composer!, { dataTransfer: { files: [file] } });
    expect(screen.getByRole("img", { name: "receipt.png" })).not.toBeNull();
    fireEvent.change(screen.getByRole("textbox", { name: "Message Caret" }), {
      target: { value: "What is in this image?" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Send message" }));

    expect(onSubmit).toHaveBeenCalledWith("What is in this image?", [file]);
    expect(screen.queryByRole("img", { name: "receipt.png" })).toBeNull();
  });

  it("rejects unsupported dropped files with an accessible error", () => {
    const file = new File(["markup"], "diagram.svg", { type: "image/svg+xml" });
    const { container } = render(
      <Composer isStreaming={false} onSubmit={vi.fn()} onStop={vi.fn()} />,
    );

    fireEvent.drop(container.querySelector(".caret-composer")!, {
      dataTransfer: { files: [file] },
    });

    expect(screen.getByRole("alert").textContent).toContain(
      "Choose JPEG, PNG, or WebP images.",
    );
    expect(screen.queryByRole("img", { name: "diagram.svg" })).toBeNull();
  });
});
