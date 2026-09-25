import { describe, expect, it } from "vitest";
import {
  HELD_ORDER_ITEM_ERROR,
  SERVER_HELD_ORDER_ITEM_ERROR,
  buildHeldOrderCartProducts,
  buildServerHeldOrderResume,
} from "./heldOrdersWorkflow";
import type { HeldOrder } from "../../shared/stores/heldOrdersStore";

const heldOrder = (items: HeldOrder["items"]): HeldOrder => ({
  id: "held_1",
  label: "Table 1",
  createdAt: "2026-01-01T10:00:00.000Z",
  items,
});

describe("heldOrdersWorkflow", () => {
  it("builds cart products from held order line ids and preserves quantity", () => {
    expect(
      buildHeldOrderCartProducts(
        heldOrder([{ id: "12::", name: "Burger", price: 10, quantity: 2 }]),
      ),
    ).toEqual([
      { id: "12::", productId: 12, name: "Burger", unitPrice: 10, price: 10 },
      { id: "12::", productId: 12, name: "Burger", unitPrice: 10, price: 10 },
    ]);
  });

  it("round-trips tax, notes, discounts, and modifiers", () => {
    const [product] = buildHeldOrderCartProducts(
      heldOrder([
        {
          id: "7::",
          productId: 7,
          name: "Burger",
          price: 100,
          quantity: 1,
          taxRate: 10,
          notes: "No onions",
          discountPct: 5,
          modifiers: [{ id: 3, name: "Extra cheese", priceDelta: 1.5 }],
        },
      ]),
    );

    expect(product).toMatchObject({
      productId: 7,
      taxRate: 10,
      notes: "No onions",
      discountPct: 5,
      modifiers: [{ id: 3, name: "Extra cheese", priceDelta: 1.5 }],
    });
  });

  it("rejects held order items without a recoverable product id", () => {
    expect(() =>
      buildHeldOrderCartProducts(
        heldOrder([
          { id: "draft-line", name: "Burger", price: 10, quantity: 1 },
        ]),
      ),
    ).toThrow(HELD_ORDER_ITEM_ERROR);
  });

  it("builds a server held-order resume request from synced item data", () => {
    expect(
      buildServerHeldOrderResume({
        id: "22",
        orderNumber: "1001",
        total: 20,
        createdAt: "2026-01-01T10:00:00.000Z",
        items: [{ menuItemId: "5", name: "Fries", quantity: 2, unitPrice: 4 }],
      }),
    ).toEqual({
      orderId: 22,
      products: [
        { id: "5", productId: 5, name: "Fries", unitPrice: 4, price: 4 },
        { id: "5", productId: 5, name: "Fries", unitPrice: 4, price: 4 },
      ],
    });
  });

  it("rejects server held orders without synced item identifiers", () => {
    expect(() =>
      buildServerHeldOrderResume({
        id: "22",
        orderNumber: "1001",
        total: 20,
        createdAt: "2026-01-01T10:00:00.000Z",
        items: [{ name: "Fries", quantity: 1 }],
      }),
    ).toThrow(SERVER_HELD_ORDER_ITEM_ERROR);
  });

  it("uses generated held-order rows directly", () => {
    const resume = buildServerHeldOrderResume({
      id: "22",
      orderNumber: "1001",
      total: 20.5,
      createdAt: "2026-01-01T10:00:00.000Z",
      items: [
        {
          menuItemId: "5",
          name: "Fries",
          quantity: 2,
          unitPrice: 4.25,
          taxRate: 10,
          notes: "No salt",
        },
      ],
    });

    expect(resume.products).toEqual([
      {
        id: "5",
        productId: 5,
        name: "Fries",
        unitPrice: 4.25,
        price: 4.25,
        taxRate: 10,
        notes: "No salt",
      },
      {
        id: "5",
        productId: 5,
        name: "Fries",
        unitPrice: 4.25,
        price: 4.25,
        taxRate: 10,
        notes: "No salt",
      },
    ]);
  });
});
