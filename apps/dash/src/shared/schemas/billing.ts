import { z } from 'zod';

export const billingSummarySchema = z.object({
  plan: z.string().default('free'),
  status: z.string().default('active'),
  nextBillingDate: z.string().optional(),
  currentUsage: z.number().optional(),
});

export const invoiceSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  number: z.string(),
  amount: z.number(),
  currency: z.string().default('USD'),
  status: z.enum(['paid', 'open', 'void', 'refunded']).catch('open'),
  createdAt: z.string(),
  pdfUrl: z.string().optional(),
});

export type BillingSummary = z.infer<typeof billingSummarySchema>;
export type Invoice = z.infer<typeof invoiceSchema>;
