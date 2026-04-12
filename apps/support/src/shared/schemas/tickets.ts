import { z } from 'zod';

export const ticketSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  subject: z.string(),
  customerName: z.string().optional(),
  customerId: z.union([z.string(), z.number()]).transform(Number).optional(),
  status: z.string(),
  priority: z.string().optional(),
  assigneeId: z.number().nullable().optional(),
  assigneeName: z.string().nullable().optional(),
  channel: z.string().optional(),
  tags: z.array(z.string()).optional(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export const ticketMessageSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  ticketId: z.number(),
  body: z.string(),
  isInternal: z.boolean().optional(),
  authorName: z.string().optional(),
  authorId: z.number().optional(),
  createdAt: z.string().optional(),
});

export type Ticket = z.infer<typeof ticketSchema>;
export type TicketMessage = z.infer<typeof ticketMessageSchema>;
