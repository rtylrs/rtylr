import type { MenuItemUpdatePayload } from '../../shared/services/api/posService';

interface ProductQuickEditDraft {
  price: number;
  stock?: number;
  available: boolean;
}

interface MenuItemUpdate {
  itemId: number;
  payload: MenuItemUpdatePayload;
}

function positiveInteger(value: unknown): number | null {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return null;
  return Math.floor(numeric);
}

function nonNegativeNumber(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  return value;
}

function nonNegativeInteger(value: number | undefined): number | undefined {
  if (value === undefined) return undefined;
  if (!Number.isFinite(value) || value < 0) return 0;
  return Math.floor(value);
}

export function buildMenuItemUpdate(
  itemId: string | number,
  draft: ProductQuickEditDraft
): MenuItemUpdate {
  const normalizedItemId = positiveInteger(itemId);
  if (!normalizedItemId) {
    throw new Error('Select a valid product before saving.');
  }

  const stock = nonNegativeInteger(draft.stock);
  return {
    itemId: normalizedItemId,
    payload: {
      price: nonNegativeNumber(draft.price),
      ...(stock === undefined ? {} : { stock }),
      available: draft.available,
    },
  };
}
