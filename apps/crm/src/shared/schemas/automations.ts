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

// -- Domain constants ---------------------------------------------------------

export const AUTOMATION_TRIGGERS = [
  'new_customer',
  'birthday',
  'visit',
  'points_threshold',
  'inactivity',
] as const;
export type AutomationTrigger = (typeof AUTOMATION_TRIGGERS)[number];

export const AUTOMATION_STATUSES = ['active', 'paused', 'draft'] as const;
export type AutomationStatus = (typeof AUTOMATION_STATUSES)[number];

// -- Type guards --------------------------------------------------------------

export function isAutomationActive(automation: Automation): boolean {
  return automation.status === 'active';
}

export function isAutomationDraft(automation: Automation): boolean {
  return automation.status === 'draft';
}
