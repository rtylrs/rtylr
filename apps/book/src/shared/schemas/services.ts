import { z } from 'zod';

export const bookServiceSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  name: z.string(),
  description: z.string().optional().nullable(),
  duration: z.number(),
  price: z.number().optional().nullable(),
  category: z.string().optional().nullable(),
  color: z.string().optional().nullable(),
  isActive: z.boolean().optional(),
  staff: z.array(z.string()).optional(),
  staffCount: z.number().optional(),
  maxBookingsPerSlot: z.number().optional().nullable(),
  bufferTime: z.number().optional().nullable(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
});

export type BookService = z.infer<typeof bookServiceSchema>;
