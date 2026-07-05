import { z } from 'zod';

export const billingSummarySchema = z
  .object({
    plan: z.string().default('free'),
    status: z.string().default('active'),
    nextBillingDate: z.string().optional(),
    next_billing_date: z.string().optional(),
    currentUsage: z.number().optional(),
    current_usage: z.number().optional(),
  })
  .transform((s) => ({
    plan: s.plan,
    status: s.status,
    nextBillingDate: s.nextBillingDate ?? s.next_billing_date ?? '',
    currentUsage: s.currentUsage ?? s.current_usage,
  }));

export const invoiceSchema = z
  .object({
    id: z.union([z.string(), z.number()]).transform(String),
    number: z.string().optional(),
    amount: z.number(),
    currency: z.string().default('USD'),
    status: z.enum(['paid', 'open', 'void', 'refunded']).catch('open'),
    createdAt: z.string().optional(),
    issued_at: z.string().optional(),
    pdfUrl: z.string().optional(),
  })
  .transform((s) => ({
    id: s.id,
    number: s.number ?? s.id,
    amount: s.amount,
    currency: s.currency,
    status: s.status,
    createdAt: s.createdAt ?? s.issued_at ?? '',
    pdfUrl: s.pdfUrl,
  }));

export type BillingSummary = z.infer<typeof billingSummarySchema>;
export type Invoice = z.infer<typeof invoiceSchema>;
