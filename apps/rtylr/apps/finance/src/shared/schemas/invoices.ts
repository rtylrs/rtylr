import { z } from 'zod';

const invoiceLineSchema = z
  .object({
    description: z.string().optional(),
    quantity: z.number().optional(),
    unitPrice: z.number().optional(),
    amount: z.number().optional(),
  })
  .passthrough();

const historyEntrySchema = z.object({
  label: z.string(),
  at: z.string(),
});

export const invoiceSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  client: z.string().default(''),
  clientEmail: z.string().optional().nullable(),
  amount: z.number().default(0),
  currency: z.string().optional().nullable(),
  paid: z.number().optional().nullable(),
  status: z.enum(['Draft', 'Sent', 'Paid', 'Overdue', 'Cancelled']).catch('Draft'),
  dueDate: z.string().default(''),
  issuedDate: z.string().default(''),
  items: z.array(invoiceLineSchema).optional().nullable(),
  notes: z.string().optional().nullable(),
  discount: z.number().optional().nullable(),
  tax: z.number().optional().nullable(),
  subtotal: z.number().optional().nullable(),
  number: z.string().optional().nullable(),
  history: z.array(historyEntrySchema).optional().nullable(),
});

export type InvoiceRaw = z.infer<typeof invoiceSchema>;
