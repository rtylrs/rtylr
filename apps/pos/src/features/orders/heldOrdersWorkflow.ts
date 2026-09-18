import type { CartProductInput } from '../../shared/utils/cart/cartProduct';
import type { HeldOrder } from '../../shared/stores/heldOrdersStore';
import type { Order as ServerHeldOrder } from '@voxire/voxie/pos/orders/v1/types';

export const HELD_ORDER_ITEM_ERROR = 'This held order contains an item that cannot be restored.';
export const SERVER_HELD_ORDER_ITEM_ERROR =
  'This server held order needs synced menu item data before it can be resumed.';

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

function nonNegativeNumber(value: unknown): number {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric < 0) return 0;
  return numeric;
}

function quantity(value: unknown): number {
  return positiveInteger(value) ?? 1;
}

function productIdFromLineId(lineId: string): number | undefined {
  return positiveInteger(lineId) ?? positiveInteger(lineId.split(':')[0]);
}

function repeatProduct(product: CartProductInput, count: number): CartProductInput[] {
  return Array.from({ length: count }, () => product);
}

export function buildHeldOrderCartProducts(order: HeldOrder): CartProductInput[] {
  return order.items.flatMap((item) => {
    const productId = productIdFromLineId(item.id);
    if (productId == null) {
      throw new Error(HELD_ORDER_ITEM_ERROR);
    }

    return repeatProduct(
      {
        id: item.id,
        productId,
        name: item.name,
        unitPrice: nonNegativeNumber(item.price),
        price: nonNegativeNumber(item.price),
      },
      quantity(item.quantity),
    );
  });
}

export function buildServerHeldOrderResume(order: ServerHeldOrder): {
  orderId: number;
  products: CartProductInput[];
} {
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
    return repeatProduct(
      {
        id: String(productId),
        productId,
        name: (item.name ?? '').trim() || 'Item',
        unitPrice: price,
        price,
      },
      quantity(item.quantity),
    );
  });

  return { orderId, products };
}
