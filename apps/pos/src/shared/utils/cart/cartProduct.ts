import type { Combo } from "../../types/pos";
import type { MenuItem } from "@voxire/voxie/pos/menu_item/v1/types";
import type { CartProductInput } from "../../stores/cartStore";

export const MENU_ITEM_CART_ERROR =
  "Select a valid menu item before adding it.";
export const COMBO_CART_ERROR = "Select a valid combo before adding it.";

export type { CartProductInput };

function positiveInteger(value: unknown): number | undefined {
  const numeric =
    typeof value === "string" ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

export function buildMenuItemCartProduct(item: MenuItem): CartProductInput {
  const productId = positiveInteger(item.id);
  if (productId == null) {
    throw new Error(MENU_ITEM_CART_ERROR);
  }

  return {
    id: item.id,
    productId,
    name: item.name ?? "",
    sku: item.sku,
    barcode: item.barcode,
    unitPrice: item.salePrice ?? 0,
    price: item.salePrice,
    taxRate: item.taxRate,
  };
}

export function buildComboCartProduct(combo: Combo): CartProductInput {
  const comboId = positiveInteger(combo.id);
  if (comboId == null) {
    throw new Error(COMBO_CART_ERROR);
  }

  return {
    id: `combo_${combo.id}`,
    productId: comboId,
    comboId,
    name: combo.name,
    unitPrice: combo.price,
    price: combo.price,
    variantName: "combo",
  };
}
