import type { SplitOrderRequest } from '@voxire/voxie/pos/orders/splits/v1/types';
import type { CartItem } from '../../shared/stores/cartStore';

export const SPLIT_UNSYNCED_ITEM_ERROR = 'Sync this order before splitting by item.';
export const SPLIT_GROUP_ERROR = 'Assign items to at least two splits before confirming.';

interface BuildSplitBillRequestInput {
  routeOrderId?: string;
  serverOrderId: number | null;
  seats: number;
  items: CartItem[];
  assignments: Record<string, number>;
}

type SplitBillRequest =
  | { localOnly: true }
  | { localOnly: false; payload: SplitOrderRequest };

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

export function buildSplitBillRequest(input: BuildSplitBillRequestInput): SplitBillRequest {
  const orderId = positiveInteger(input.serverOrderId) ?? positiveInteger(input.routeOrderId);
  if (orderId == null) {
    return { localOnly: true };
  }

  const seats = positiveInteger(input.seats);
  if (seats == null || seats < 2) {
    throw new Error(SPLIT_GROUP_ERROR);
  }

  const splitGroups: number[][] = Array.from({ length: seats }, (_, seatIdx) => {
    const group = input.items.filter((item) => (input.assignments[item.id] ?? 0) === seatIdx);

    return group.map((item) => {
      const itemId = positiveInteger(item.id);
      if (itemId == null) {
        throw new Error(SPLIT_UNSYNCED_ITEM_ERROR);
      }
      return itemId;
    });
  }).filter((group) => group.length > 0);

  if (splitGroups.length < 2) {
    throw new Error(SPLIT_GROUP_ERROR);
  }

  return {
    localOnly: false,
    payload: {
      orderId: String(orderId),
      splits: splitGroups.map((itemIds) => ({ itemIds: itemIds.map(String) })),
    },
  };
}
