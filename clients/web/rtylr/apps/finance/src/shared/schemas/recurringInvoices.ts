import { z } from 'zod';

/**
 * Mirrors finance.recurring_invoice.v1.RecurringInvoice. id/ownerId/
 * templateInvoiceId are onekit int64-as-string.
 */
export const recurringInvoiceSchema = z
  .object({
    id: z.string(),
    ownerId: z.string().optional(),
    templateInvoiceId: z.string().optional(),
    frequency: z.string(),
    nextInvoiceDate: z.string().optional().nullable(),
    endDate: z.string().optional().nullable(),
    isActive: z.boolean(),
    createdAt: z.string().optional().nullable(),
  })
  .passthrough();

export type RecurringInvoiceRaw = z.infer<typeof recurringInvoiceSchema>;
