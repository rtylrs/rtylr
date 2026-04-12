import { z } from 'zod';

export const onboardingTaskSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  task: z.string(),
  category: z
    .enum(['Paperwork', 'Access', 'Training', 'Meet the team', 'Equipment'])
    .catch('Paperwork'),
  completed: z.boolean(),
  dueDate: z.string().optional(),
});

export const offboardingTaskSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  task: z.string(),
  category: z.enum(['Knowledge transfer', 'Access', 'Equipment', 'Paperwork']).catch('Paperwork'),
  completed: z.boolean(),
  dueDate: z.string().optional(),
});

export type OnboardingTask = z.infer<typeof onboardingTaskSchema>;
export type OffboardingTask = z.infer<typeof offboardingTaskSchema>;
