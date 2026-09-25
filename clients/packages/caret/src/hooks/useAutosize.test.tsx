import { render } from "@testing-library/react";
import { useRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { useAutosize } from "./useAutosize";

function Harness({ value }: { value: string }): React.JSX.Element {
  const ref = useRef<HTMLTextAreaElement>(null);
  useAutosize(ref, value, 6);
  return <textarea ref={ref} aria-label="composer" />;
}

describe("useAutosize", () => {
  it("uses the textarea computed line height and padding for min and max heights", () => {
    vi.spyOn(window, "getComputedStyle").mockReturnValue({
      lineHeight: "20px",
      paddingTop: "6px",
      paddingBottom: "6px",
    } as CSSStyleDeclaration);

    const view = render(<Harness value="" />);
    const textarea = view.getByLabelText("composer") as HTMLTextAreaElement;
    expect(textarea.style.height).toBe("32px");

    Object.defineProperty(textarea, "scrollHeight", {
      configurable: true,
      value: 1_000,
    });
    view.rerender(<Harness value="six lines of text" />);
    expect(textarea.style.height).toBe("132px");
  });
});
