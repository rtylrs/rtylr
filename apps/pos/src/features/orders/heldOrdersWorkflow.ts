import type { CartProductInput } from "../../shared/stores/cartStore";
import type { HeldOrder } from "../../shared/stores/heldOrdersStore";
import type { Order as ServerHeldOrder } from "@voxire/voxie/pos/orders/v1/types";

export const HELD_ORDER_ITEM_ERROR =
  "This held order contains an item that cannot be restored.";
export const SERVER_HELD_ORDER_ITEM_ERROR =
  "This server held order needs synced menu item data before it can be resumed.";

function positiveInteger(value: unknown): number | undefined {
  const numeric =
    typeof value === "string" ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

function nonNegativeNumber(value: unknown): number {
  const numeric = Number(value);
  return Number.isFinite(numeric) && numeric >= 0 ? numeric : 0;
}

function quantity(value: unknown): number {
  return positiveInteger(value) ?? 1;
}

function productIdFromLineId(lineId: string): number | undefined {
  const trimmed = lineId.trim();
  const comboId = positiveInteger(trimmed.match(/^combo_(\d+)/)?.[1]);
  return (
    comboId ??
    positiveInteger(trimmed) ??
    positiveInteger(trimmed.split(":")[0])
  );
}

function repeatProduct(
  product: CartProductInput,
  count: number,
): CartProductInput[] {
  return Array.from({ length: count }, () => ({ ...product }));
}

function heldProduct(item: HeldOrder["items"][number]): CartProductInput {
  const productId =
    positiveInteger(item.productId) ?? productIdFromLineId(item.id);
  if (productId == null) {
    throw new Error(HELD_ORDER_ITEM_ERROR);
  }

  const product: CartProductInput = {
    id: item.id,
    productId,
    name: item.name,
    unitPrice: nonNegativeNumber(item.price),
    price: nonNegativeNumber(item.price),
  };
  if (item.sku) product.sku = item.sku;
  if (item.barcode) product.barcode = item.barcode;
  if (item.notes?.trim()) product.notes = item.notes;
  if (nonNegativeNumber(item.taxRate) !== 0)
    product.taxRate = nonNegativeNumber(item.taxRate);
  if (item.modifiers?.length) product.modifiers = item.modifiers;
  if (positiveInteger(item.comboId))
    product.comboId = positiveInteger(item.comboId);
  if (item.variantName) product.variantName = item.variantName;
  if (nonNegativeNumber(item.discountPct) !== 0)
    product.discountPct = nonNegativeNumber(item.discountPct);
  if (nonNegativeNumber(item.discountAmount) !== 0)
    product.discountAmount = nonNegativeNumber(item.discountAmount);
  return product;
}

export function buildHeldOrderCartProducts(
  order: HeldOrder,
): CartProductInput[] {
  return order.items.flatMap((item) =>
    repeatProduct(heldProduct(item), quantity(item.quantity)),
  );
}

export interface ServerHeldOrderResume {
  orderId: number;
  products: CartProductInput[];
  customerId?: number;
  customerName?: string;
  notes?: string;
  kitchenNotes?: string;
  tableId?: number;
  locationId?: number;
  orderType?: "dine_in" | "takeaway" | "delivery" | "online";
  discountAmount?: number;
}

function orderType(value: unknown): ServerHeldOrderResume["orderType"] {
  return value === "dine_in" || value === "delivery" || value === "online"
    ? value
    : "takeaway";
}

export function buildServerHeldOrderResume(
  order: ServerHeldOrder,
): ServerHeldOrderResume {
  const orderId = positiveInteger(order.id);
  if (orderId == null) {
    throw new Error(SERVER_HELD_ORDER_ITEM_ERROR);
  }

  const products = (order.items ?? []).flatMap((item) => {
    const productId = positiveInteger(item.menuItemId);
    if (productId == null) {
      throw new Error(SERVER_HELD_ORDER_ITEM_ERROR);
    }

    const price = nonNegativeNumber(item.unitPrice);
    const serverItemId = positiveInteger(item.id);
    const lineId = serverItemId ? String(serverItemId) : String(productId);
    const product: CartProductInput = {
      id: lineId,
      productId,
      name: (item.name ?? "").trim() || "Item",
      unitPrice: price,
      price,
    };
    if (serverItemId) product.lineId = lineId;
    if (item.sku) product.sku = item.sku;
    if (item.notes?.trim()) product.notes = item.notes;
    if (nonNegativeNumber(item.taxRate) !== 0)
      product.taxRate = nonNegativeNumber(item.taxRate);
    if (nonNegativeNumber(item.discountPct) !== 0)
      product.discountPct = nonNegativeNumber(item.discountPct);
    if (nonNegativeNumber(item.discountAmount) !== 0)
      product.discountAmount = nonNegativeNumber(item.discountAmount);
    return repeatProduct(product, quantity(item.quantity));
  });

  const result: ServerHeldOrderResume = { orderId, products };
  const customerId = positiveInteger(order.customerId);
  const customerName = order.customerName?.trim();
  const notes = order.notes?.trim();
  const kitchenNotes = order.kitchenNotes?.trim();
  const tableId = positiveInteger(order.tableId);
  const locationId = positiveInteger(order.locationId);
  const discountAmount = nonNegativeNumber(order.discountAmount);
  if (customerId != null) result.customerId = customerId;
  if (customerName) result.customerName = customerName;
  if (notes) result.notes = notes;
  if (kitchenNotes) result.kitchenNotes = kitchenNotes;
  if (tableId != null) result.tableId = tableId;
  if (locationId != null) result.locationId = locationId;
  if (discountAmount > 0) result.discountAmount = discountAmount;
  if (order.type) result.orderType = orderType(order.type);
  return result;
}
