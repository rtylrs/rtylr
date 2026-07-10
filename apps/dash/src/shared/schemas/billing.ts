import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

export const billingSummarySchema = z
  .object({
    plan: z.string().default('free'),
    status: z.string().default('active'),
    nextBillingDate: z.string().optional(),
    next_billing_date: z.string().optional(),
    currentUsage: finiteNumber.optional(),
    current_usage: finiteNumber.optional(),
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
    amount: finiteNumber,
    currency: z.string().default('USD'),
    status: z.enum(['paid', 'open', 'void', 'refunded']).catch('open'),
    createdAt: z.string().optional(),
    created_at: z.string().optional(),
    issued_at: z.string().optional(),
    pdfUrl: z.string().optional(),
    pdf_url: z.string().optional(),
  })
  .transform((s) => ({
    id: s.id,
    number: s.number ?? s.id,
    amount: s.amount,
    currency: s.currency,
    status: s.status,
    createdAt: s.createdAt ?? s.created_at ?? s.issued_at ?? '',
    pdfUrl: s.pdfUrl ?? s.pdf_url,
  }));

export type BillingSummary = z.infer<typeof billingSummarySchema>;
export type Invoice = z.infer<typeof invoiceSchema>;
