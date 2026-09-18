import { z } from 'zod';

function record(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null;
  return value as Record<string, unknown>;
}

function text(value: unknown): string | undefined {
  if (typeof value === 'string') {
    const trimmed = value.trim();
    return trimmed || undefined;
  }
  if (typeof value === 'number' && Number.isFinite(value)) return String(value);
  return undefined;
}

function number(value: unknown): number {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

const paymentEntrySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  method: z.enum(['cash', 'card', 'gift_card', 'other']).catch('other'),
  amount: z.coerce.number().catch(0),
  reference: z.string().optional().nullable(),
  tipAmount: z.coerce.number().optional().nullable(),
  createdAt: z.string().default(() => new Date().toISOString()),
});

const receiptItemSchema = z.object({
  name: z.string().default('Item'),
  qty: z.coerce.number().catch(1),
  price: z.coerce.number().catch(0),
});

const receiptObjectSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  orderId: z.union([z.string(), z.number()]).transform(String),
  total: z.coerce.number().catch(0),
  customerName: z.string().optional().nullable(),
  emailed: z.boolean().optional().nullable(),
  printedAt: z.string().default(() => new Date().toISOString()),
  items: z.array(receiptItemSchema).default([]),
  payments: z.array(paymentEntrySchema).default([]),
});

export const receiptSchema = z.preprocess((value) => {
  const raw = record(value);
  if (!raw) return value;

  const id = text(raw.id) ?? '';
  const printedAt = text(raw.printedAt ?? raw.printed_at ?? raw.paidAt ?? raw.paid_at);
  const total = number(raw.total);
  const paymentMethod = text(raw.paymentMethod ?? raw.payment_method);
  const existingPayments = Array.isArray(raw.payments) ? raw.payments : undefined;

  return {
    ...raw,
    id,
    orderId:
      raw.orderId ??
      raw.order_id ??
      raw.orderNumber ??
      raw.order_number ??
      id,
    total,
    customerName: raw.customerName ?? raw.customer_name,
    printedAt: printedAt ?? raw.printedAt,
    items: Array.isArray(raw.items) ? raw.items : [],
    payments:
      existingPayments ??
      (paymentMethod
        ? [
            {
              id: `payment-${id}`,
              method: paymentMethod,
              amount: total,
              createdAt: printedAt ?? new Date().toISOString(),
            },
          ]
        : []),
  };
}, receiptObjectSchema);

export type ReceiptRaw = z.infer<typeof receiptSchema>;
