import { z } from 'zod';

export const customerSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string().default('Customer'),
  phone: z.string().optional().nullable(),
  email: z.string().optional().nullable(),
  loyaltyPoints: z.number().optional().nullable(),
  lastVisit: z.string().optional().nullable(),
});

export type CustomerRaw = z.infer<typeof customerSchema>;
