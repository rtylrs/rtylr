import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import MenuItem from "./index";
import type { ItemDisplayConfig } from "../../types";

const displayConfig: ItemDisplayConfig = {
  layout: "minimal",
  columns: 2,
  imagePosition: "none",
  imageSize: "sm",
  showDescription: false,
  showBadges: false,
  cardRadius: "none",
};

describe("MenuItem", () => {
  it("shows unavailable state without an image and formats the item currency", () => {
    render(
      <MenuItem
        item={{
          id: "1",
          name: "Unavailable item",
          price: 12,
          currency: "EUR",
          available: false,
        }}
        displayConfig={displayConfig}
      />,
    );

    expect(screen.getByText("Unavailable")).toBeTruthy();
    expect(screen.getByText("€12.00")).toBeTruthy();
  });

  it("does not throw for a malformed currency code", () => {
    expect(() =>
      render(
        <MenuItem
          item={{ id: "1", name: "Item", price: 12, currency: "US" }}
          displayConfig={displayConfig}
        />,
      ),
    ).not.toThrow();
    expect(screen.getByText("US 12.00")).toBeTruthy();
  });
});
