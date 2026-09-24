import { describe, expect, it } from "vitest";
import { buildComboCartProduct, buildMenuItemCartProduct } from "./cartProduct";

describe("cartProduct", () => {
  it("builds a cart product from a valid menu item", () => {
    expect(
      buildMenuItemCartProduct({
        id: "17",
        name: "Burger",
        salePrice: 12.5,
        category: "Mains",
        sku: "BRG",
        taxRate: 0.11,
        barcode: "123",
      }),
    ).toEqual({
      id: "17",
      productId: 17,
      name: "Burger",
      sku: "BRG",
      barcode: "123",
      unitPrice: 12.5,
      price: 12.5,
      taxRate: 0.11,
    });
  });

  it("rejects menu items without a valid product ID", () => {
    expect(() =>
      buildMenuItemCartProduct({
        id: "local-1",
        name: "Burger",
        salePrice: 12,
        category: "Mains",
      }),
    ).toThrow("Select a valid menu item before adding it.");
  });

  it("builds a cart product from a valid combo", () => {
    expect(
      buildComboCartProduct({
        id: "3",
        name: "Lunch Combo",
        price: 20,
        items: [],
      }),
    ).toEqual({
      id: "combo_3",
      productId: 3,
      comboId: 3,
      name: "Lunch Combo",
      unitPrice: 20,
      price: 20,
      variantName: "combo",
    });
  });

  it("rejects combos without a valid combo ID", () => {
    expect(() =>
      buildComboCartProduct({
        id: "draft-combo",
        name: "Draft Combo",
        price: 20,
        items: [],
      }),
    ).toThrow("Select a valid combo before adding it.");
  });
});
