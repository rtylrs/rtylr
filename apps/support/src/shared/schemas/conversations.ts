import { z } from 'zod';

export const conversationSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  subject: z.string().optional(),
  customerName: z.string().optional(),
  customerId: z.number().optional(),
  channel: z.string().optional(),
  status: z.string().optional(),
  assigneeId: z.number().nullable().optional(),
  assigneeName: z.string().nullable().optional(),
  unread: z.boolean().optional(),
  lastMessage: z.string().optional(),
  lastMessageAt: z.string().optional(),
  createdAt: z.string().optional(),
});

export type Conversation = z.infer<typeof conversationSchema>;
