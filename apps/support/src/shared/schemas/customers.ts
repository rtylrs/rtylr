import { z } from 'zod';

export const supportCustomerSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  email: z.string().optional(),
  phone: z.string().optional(),
  ticketCount: z.number().optional(),
  createdAt: z.string().optional(),
  lastContactedAt: z.string().optional(),
  tags: z.array(z.string()).optional(),
});

export type SupportCustomer = z.infer<typeof supportCustomerSchema>;
