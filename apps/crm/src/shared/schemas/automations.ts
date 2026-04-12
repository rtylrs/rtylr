import { z } from 'zod';

export const automationNodeSchema = z.object({
  id: z.string(),
  type: z.string(),
  label: z.string(),
  x: z.number(),
  y: z.number(),
});

export const automationSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  description: z.string(),
  trigger: z.string(),
  status: z.enum(['active', 'paused', 'draft']).catch('draft'),
  enrolled: z.number(),
  completed: z.number(),
  createdAt: z.string(),
  nodes: z.array(automationNodeSchema).optional(),
});

export type Automation = z.infer<typeof automationSchema>;
