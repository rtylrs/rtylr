import { z } from 'zod';
import { kdsOrderSchema } from '../../shared/utils/schemas/kdsOrders';
import type { KDSOrder } from './useKDSOrders';

export const KDS_ITEM_ID_ERROR = 'Select a valid KDS item before updating status.';

const positiveIntegerSchema = z.coerce.number().int().positive();

export function parsePositiveInteger(value: unknown): number | undefined {
  const parsed = positiveIntegerSchema.safeParse(value);
  return parsed.success ? parsed.data : undefined;
}

export function mapKdsOrders(raw: unknown): KDSOrder[] {
  const parsed = z.array(kdsOrderSchema).parse(Array.isArray(raw) ? raw : []);

  return parsed.map((o) => ({
    id: o.id,
    number: o.number ?? o.id,
    station: o.station,
    items: o.items.map((it) => ({
      name: it.name,
      qty: it.qty,
      course: it.course ?? undefined,
      notes: it.notes ?? undefined,
    })),
    status: o.status,
    createdAt: o.createdAt,
    elapsedSeconds: o.elapsedSeconds,
  }));
}

export function buildKdsStatusUpdate(
  itemId: string,
  status: KDSOrder['status'],
): { itemId: number; status: KDSOrder['status'] } {
  const parsedItemId = parsePositiveInteger(itemId);
  if (parsedItemId == null) {
    throw new Error(KDS_ITEM_ID_ERROR);
  }

  return { itemId: parsedItemId, status };
}
