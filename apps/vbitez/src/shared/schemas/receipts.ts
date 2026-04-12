import { z } from 'zod';

const paymentEntrySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  method: z.enum(['cash', 'card', 'gift_card', 'other']).catch('other'),
  amount: z.number().default(0),
  reference: z.string().optional().nullable(),
  tipAmount: z.number().optional().nullable(),
  createdAt: z.string().default(() => new Date().toISOString()),
});

const receiptItemSchema = z.object({
  name: z.string().default('Item'),
  qty: z.number().default(1),
  price: z.number().default(0),
});

export const receiptSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  orderId: z.union([z.string(), z.number()]).transform(String),
  total: z.number().default(0),
  customerName: z.string().optional().nullable(),
  emailed: z.boolean().optional().nullable(),
  printedAt: z.string().default(() => new Date().toISOString()),
  items: z.array(receiptItemSchema).default([]),
  payments: z.array(paymentEntrySchema).default([]),
});

export type ReceiptRaw = z.infer<typeof receiptSchema>;
