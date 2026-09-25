import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MenuLayoutCore from "./index";
import type { MenuCategoryData } from "../../components/MenuCategory";
import type { MenuLayoutConfig } from "../../types";

const categories: MenuCategoryData[] = [
  {
    id: "coffee",
    name: "Coffee",
    items: [{ id: "1", name: "Espresso", price: 3 }],
  },
];

const config: Partial<MenuLayoutConfig> = {
  pages: { gallery: true },
};

describe("MenuLayoutCore", () => {
  it("falls back to the menu when the active page is disabled by a config update", () => {
    vi.spyOn(window, "scrollTo").mockImplementation(() => undefined);
    const view = render(
      <MenuLayoutCore
        business={{ name: "Cafe" }}
        categories={categories}
        config={config}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Gallery" }));
    expect(screen.getByText("No photos yet.")).toBeTruthy();

    view.rerender(
      <MenuLayoutCore
        business={{ name: "Cafe" }}
        categories={categories}
        config={{ pages: { gallery: false } }}
      />,
    );

    expect(screen.getByRole("heading", { name: "Coffee" })).toBeTruthy();
    expect(screen.queryByText("No photos yet.")).toBeNull();
  });
});
