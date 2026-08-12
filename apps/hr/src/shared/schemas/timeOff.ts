import { z } from 'zod';

export const timeOffRequestSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  startDate: z.string(),
  endDate: z.string(),
  reason: z.string(),
  status: z
    .string()
    .transform((s) => {
      const lower = (s || '').toLowerCase();
      if (lower === 'approved') return 'Approved';
      if (lower === 'rejected') return 'Rejected';
      return 'Pending';
    })
    .catch('Pending'),
});

export const timeOffEntrySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  type: z.enum(['Vacation', 'Sick', 'Personal', 'Bereavement', 'Unpaid']).catch('Vacation'),
  startDate: z.string(),
  endDate: z.string(),
  reason: z.string(),
  status: z
    .string()
    .transform((s) => {
      const lower = (s || '').toLowerCase();
      if (lower === 'approved') return 'Approved';
      if (lower === 'rejected') return 'Rejected';
      return 'Pending';
    })
    .catch('Pending'),
  note: z.string().optional(),
  createdAt: z.string(),
});

export type TimeOffRequest = z.infer<typeof timeOffRequestSchema>;
export type TimeOffEntry = z.infer<typeof timeOffEntrySchema>;
