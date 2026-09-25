import { beforeEach, describe, expect, it, vi } from "vitest";

const restPost = vi.hoisted(() => vi.fn());
const restPut = vi.hoisted(() => vi.fn());
const unwrap = vi.hoisted(() => vi.fn());

vi.mock("@voxire/api", () => ({
  createRestRouteClient: () => ({ post: restPost, put: restPut }),
  unwrap,
}));

import {
  buildOrderCompatPayload,
  buildResumeOrderPayload,
  createOrderCompat,
  hasComboItems,
  persistResumedOrder,
} from "./orderCompatWorkflow";
import type { CartItem } from "../../shared/stores/cartStore";

const item = (overrides: Partial<CartItem> = {}): CartItem => ({
  lineId: "7::",
  id: "7::",
  productId: 7,
  name: "Burger",
  unitPrice: 10,
  price: 10,
  quantity: 2,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: "pending",
  ...overrides,
});

describe("order compatibility workflow", () => {
  beforeEach(() => {
    restPost.mockReset();
    restPut.mockReset();
    unwrap.mockReset();
  });

  it("keeps combo identity separate from menu item identity", () => {
    const menu = item();
    const combo = item({
      lineId: "combo:7",
      id: "combo:7",
      comboId: 7,
      variantName: "combo",
    });

    expect(hasComboItems([menu])).toBe(false);
    expect(hasComboItems([combo])).toBe(true);
    expect(
      buildOrderCompatPayload({ items: [combo], sessionId: 42 }).items[0],
    ).toMatchObject({
      comboId: 7,
      quantity: 2,
    });
    expect(
      buildOrderCompatPayload({ items: [combo], sessionId: 42 }).items[0],
    ).not.toHaveProperty("menuItemId");
  });

  it("serializes mixed menu and combo lines without numeric-id collisions", () => {
    const payload = buildOrderCompatPayload({
      items: [
        item(),
        item({
          lineId: "combo_7",
          id: "combo_7",
          comboId: 7,
          variantName: "combo",
        }),
      ],
      sessionId: 42,
    });

    expect(payload.items).toEqual([
      expect.objectContaining({ menuItemId: 7 }),
      expect.objectContaining({ comboId: 7 }),
    ]);
    expect(payload.items[1]).not.toHaveProperty("menuItemId");
  });

  it("builds a numeric, tenant-scoped resume payload with notes and tax", () => {
    const payload = buildResumeOrderPayload({
      orderId: 22,
      items: [item({ notes: "No onions", taxRate: 10 })],
      locationId: 3,
      customerId: 8,
    });

    expect(payload.orderId).toBe(22);
    expect(payload.locationId).toBe(3);
    expect(payload.customerId).toBe(8);
    expect(payload.items[0]).toMatchObject({
      menuItemId: 7,
      taxRate: 10,
      notes: "No onions",
      discountAmount: 0,
    });
  });

  it("posts combo creates and resumes through the compatibility boundary", async () => {
    unwrap.mockResolvedValueOnce({ orderId: 7, orderNumber: "A-7" });
    const payload = buildOrderCompatPayload({ items: [item()], sessionId: 42 });
    await expect(createOrderCompat(payload)).resolves.toMatchObject({
      orderId: "7",
    });
    expect(restPost).toHaveBeenCalledWith("/pos/orders/compat", payload);

    unwrap.mockResolvedValueOnce({ success: true, orderId: 7 });
    const resume = buildResumeOrderPayload({ orderId: 7, items: [item()] });
    await expect(persistResumedOrder(resume)).resolves.toBeUndefined();
    expect(restPut).toHaveBeenCalledWith("/pos/orders/7/resume", resume);
  });
});
