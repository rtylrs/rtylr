import { z } from 'zod';

export const reviewSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  reviewerId: z.number(),
  reviewer: z.string(),
  period: z.string(),
  status: z.enum(['Draft', 'In Progress', 'Submitted', 'Completed']).catch('Draft'),
  overall: z.number(),
  goals: z.number(),
  values: z.number(),
  skills: z.number(),
  notes: z.string(),
  date: z.string(),
});

export const oneOnOneSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  managerId: z.number(),
  manager: z.string(),
  reportId: z.number(),
  report: z.string(),
  scheduledAt: z.string(),
  duration: z.number(),
  status: z.enum(['Scheduled', 'Completed', 'Cancelled']).catch('Scheduled'),
  topics: z.array(z.string()).default([]),
  notes: z.string().optional(),
});

export type Review = z.infer<typeof reviewSchema>;
export type OneOnOne = z.infer<typeof oneOnOneSchema>;
