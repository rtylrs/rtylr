import { z } from 'zod';

export const createCycleSchema = z.object({
  name: z.string().min(1),
  periodStart: z.string().min(1),
  periodEnd: z.string().min(1),
  reviewTypes: z.array(z.enum(['self', 'manager', 'peer'])).min(1),
});

export type CreateCycleForm = z.infer<typeof createCycleSchema>;

export const goalFormSchema = z.object({
  title: z.string().min(1),
  description: z.string().default(''),
  weight: z.number().min(0).max(100).optional(),
});

export type GoalFormData = z.infer<typeof goalFormSchema>;

export const myReviewSchema = z.object({
  overallRating: z.string().min(1),
  narrative: z.string().min(50),
});

export type MyReviewForm = z.infer<typeof myReviewSchema>;
