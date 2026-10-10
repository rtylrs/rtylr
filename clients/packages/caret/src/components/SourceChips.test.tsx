import { describe, expect, it } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import SourceChips from "./SourceChips";

const sources = Array.from({ length: 6 }, (_, index) => ({
  title: `Story ${index}`,
  url: `https://site${index}.test/`,
  host: `site${index}.test`,
}));

describe("SourceChips", () => {
  it("shows the first few sources and expands the rest", () => {
    render(<SourceChips sources={sources} />);

    expect(screen.getAllByRole("link")).toHaveLength(4);
    expect(
      screen.getByRole("link", { name: /Story 0/ }).getAttribute("href"),
    ).toBe("https://site0.test/");

    fireEvent.click(screen.getByRole("button", { name: "+2 more" }));
    expect(screen.getAllByRole("link")).toHaveLength(6);
  });

  it("renders nothing without sources", () => {
    const { container } = render(<SourceChips sources={[]} />);
    expect(container.childElementCount).toBe(0);
  });
});
