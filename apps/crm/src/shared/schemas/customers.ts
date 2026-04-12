import { z } from 'zod';

export const customerSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  email: z.string(),
  phone: z.string(),
  totalSpent: z.number(),
  visits: z.number(),
  loyaltyPoints: z.number(),
  lastVisit: z.string(),
  initials: z.string(),
  tags: z.array(z.string()).optional(),
  notes: z.string().optional(),
  birthday: z.string().optional(),
  consent: z
    .object({
      email: z.boolean(),
      sms: z.boolean(),
      marketing: z.boolean(),
    })
    .optional(),
  createdAt: z.string().optional(),
  referredBy: z.number().nullable().optional(),
});

export const timelineEntrySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  type: z.enum(['visit', 'purchase', 'email', 'sms', 'note', 'points', 'reward']).catch('visit'),
  title: z.string(),
  description: z.string().optional(),
  at: z.string(),
  meta: z.record(z.string(), z.unknown()).optional(),
});

export type Customer = z.infer<typeof customerSchema>;
export type TimelineEntry = z.infer<typeof timelineEntrySchema>;
