import { z } from 'zod';

export const waitlistEntrySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  customerId: z.union([z.string(), z.number()]).transform(String).optional().nullable(),
  serviceId: z.union([z.string(), z.number()]).transform(String).optional().nullable(),
  customerName: z.string().optional(),
  customerEmail: z.string().optional().nullable(),
  customerPhone: z.string().optional().nullable(),
  serviceName: z.string().optional().nullable(),
  status: z.enum(['waiting', 'notified', 'booked', 'expired', 'removed']).catch('waiting'),
  preferredDate: z.string().optional().nullable(),
  preferredTime: z.string().optional().nullable(),
  notes: z.string().optional().nullable(),
  createdAt: z.string().optional(),
  notifiedAt: z.string().optional().nullable(),
});

export type WaitlistEntry = z.infer<typeof waitlistEntrySchema>;
