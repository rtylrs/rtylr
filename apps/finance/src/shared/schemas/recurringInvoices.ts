import { z } from 'zod';

/**
 * Mirrors the DEPLOYED onion `RecurringInvoicePayload`
 * (recurring_invoices module, handlers.go). The @voxire/api interceptor
 * converts the backend snake_case keys to camelCase before this schema runs,
 * so all fields are camelCase here.
 *
 * NOTE: the proto in @voxire/voxie/finance/v1/recurring_invoices declares a
 * richer shape (customerId, lines, subtotal, ...) that the backend does not yet
 * implement. The deployed wire shape below is the source of truth.
 */
export const recurringInvoiceSchema = z
  .object({
    id: z.number(),
    ownerId: z.number().optional(),
    templateInvoiceId: z.number().optional(),
    frequency: z.string(),
    nextInvoiceDate: z.string().optional().nullable(),
    endDate: z.string().optional().nullable(),
    isActive: z.boolean(),
    createdAt: z.string().optional().nullable(),
  })
  .passthrough();

export type RecurringInvoiceRaw = z.infer<typeof recurringInvoiceSchema>;
