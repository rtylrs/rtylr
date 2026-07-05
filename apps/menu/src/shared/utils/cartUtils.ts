import type { CartItem } from '../stores/cartStore';

/**
 * Computes the subtotal for a single cart line item.
 */
export function lineSubtotal(item: CartItem): number {
  return item.unitPrice * item.quantity;
}

/**
 * Computes the grand total for a collection of cart items.
 */
export function cartTotal(items: CartItem[]): number {
  return items.reduce((sum, item) => sum + lineSubtotal(item), 0);
}

/**
 * Returns the total number of individual units across all cart items.
 */
export function cartItemCount(items: CartItem[]): number {
  return items.reduce((sum, item) => sum + item.quantity, 0);
}

/**
 * Returns true when the cart has at least one item.
 */
export function cartIsEmpty(items: CartItem[]): boolean {
  return items.length === 0;
}

/**
 * Finds a cart item by its menu item ID, returning undefined if not found.
 */
export function findCartItem(items: CartItem[], menuItemId: number): CartItem | undefined {
  return items.find((i) => i.menuItemId === menuItemId);
}

/**
 * Builds an order-line payload array suitable for the public orders OneKit API.
 */
export function toOrderLines(
  items: CartItem[]
): { menu_item_id: number; name: string; quantity: number; unit_price: number; notes: string }[] {
  return items.map((i) => ({
    menu_item_id: i.menuItemId,
    name: i.name,
    quantity: i.quantity,
    unit_price: i.unitPrice,
    notes: i.notes,
  }));
}
