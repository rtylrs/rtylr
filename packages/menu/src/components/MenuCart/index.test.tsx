import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import MenuCart from "./index";

describe("MenuCart", () => {
  it("uses the cart item currency for line and subtotal amounts", () => {
    render(
      <MenuCart
        items={[
          {
            id: "1",
            name: "Espresso",
            price: 12,
            currency: "EUR",
            quantity: 2,
          },
        ]}
        isOpen
        onClose={vi.fn()}
        onQuantityChange={vi.fn()}
        onCheckout={vi.fn()}
      />,
    );

    expect(screen.getByText("€12.00")).toBeTruthy();
    expect(screen.getByText("€24.00")).toBeTruthy();
    expect(
      screen.getByRole("button", { name: /Place Order · €24\.00/ }),
    ).toBeTruthy();
  });
});
