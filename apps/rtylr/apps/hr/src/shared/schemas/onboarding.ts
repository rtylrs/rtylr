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

// The business-wide summary payload carries int64 ids/timestamps as strings
// on the wire (protobuf JSON convention) - transform to number so callers
// don't re-derive this at every call site. Zero-valued numeric fields are
// dropped entirely by the backend's `omitempty` tags rather than sent as 0,
// so every count defaults to 0 instead of being required.
export const employeeChecklistSummarySchema = z.object({
  employeeId: z.union([z.string(), z.number()]).transform(Number),
  checklistId: z.union([z.string(), z.number()]).transform(Number),
  templateId: z.union([z.string(), z.number()]).transform(Number),
  status: z.string(),
  totalItems: z.number().optional().default(0),
  completedItems: z.number().optional().default(0),
  completionPercentage: z.number().optional().default(0),
  overdueItemsCount: z.number().optional().default(0),
  triggeredAt: z.union([z.string(), z.number()]).transform(Number),
  updatedAt: z.union([z.string(), z.number()]).transform(Number),
});

export const employeeChecklistSummaryListSchema = z.object({
  summaries: z.array(employeeChecklistSummarySchema),
  total: z.number(),
  limit: z.number(),
  offset: z.number(),
});

export type EmployeeChecklistSummary = z.infer<typeof employeeChecklistSummarySchema>;
export type EmployeeChecklistSummaryList = z.infer<typeof employeeChecklistSummaryListSchema>;
