/**
 * Analytics event helpers for the Menu app.
 * These utilities build typed event payloads for analytics tracking.
 */

export interface MenuAnalyticsEvent {
  event: string;
  properties: Record<string, unknown>;
}

/** Build a view event for a menu page. */
export function menuViewEvent(menuSlug: string): MenuAnalyticsEvent {
  return {
    event: 'menu_viewed',
    properties: { menu_slug: menuSlug, timestamp: Date.now() },
  };
}

/** Build an add-to-cart event. */
export function addToCartEvent(
  itemId: number,
  itemName: string,
  price: number,
  quantity: number
): MenuAnalyticsEvent {
  return {
    event: 'add_to_cart',
    properties: { item_id: itemId, item_name: itemName, price, quantity },
  };
}

/** Build a remove-from-cart event. */
export function removeFromCartEvent(itemId: number, itemName: string): MenuAnalyticsEvent {
  return {
    event: 'remove_from_cart',
    properties: { item_id: itemId, item_name: itemName },
  };
}

/** Build a checkout initiated event with the cart subtotal. */
export function checkoutInitiatedEvent(subtotal: number, itemCount: number): MenuAnalyticsEvent {
  return {
    event: 'checkout_initiated',
    properties: { subtotal, item_count: itemCount },
  };
}

/** Build an order placed event. */
export function orderPlacedEvent(orderId: string, total: number): MenuAnalyticsEvent {
  return {
    event: 'order_placed',
    properties: { order_id: orderId, total },
  };
}

/** Build a search event. */
export function searchEvent(query: string, resultCount: number): MenuAnalyticsEvent {
  return {
    event: 'menu_searched',
    properties: { query: query.slice(0, 200), result_count: resultCount },
  };
}
