import { z } from 'zod';

export const trainingCourseSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  title: z.string(),
  category: z.enum(['Compliance', 'Safety', 'Skills', 'Leadership']).catch('Compliance'),
  durationMin: z.number(),
  modules: z.number(),
  required: z.boolean(),
});

export const trainingProgressSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  courseId: z.number(),
  courseName: z.string(),
  progress: z.number(),
  completedAt: z.string().optional(),
});

export type TrainingCourse = z.infer<typeof trainingCourseSchema>;
export type TrainingProgress = z.infer<typeof trainingProgressSchema>;
