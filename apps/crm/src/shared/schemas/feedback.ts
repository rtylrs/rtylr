import { z } from 'zod';

export const feedbackItemSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  customer: z.string(),
  rating: z.number(),
  nps: z.number(),
  comment: z.string(),
  createdAt: z.string(),
  source: z.string(),
});

export type FeedbackItem = z.infer<typeof feedbackItemSchema>;
