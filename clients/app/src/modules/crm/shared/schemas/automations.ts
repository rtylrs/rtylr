import { z } from 'zod';

export const automationNodeSchema = z.object({
  id: z.string(),
  type: z.string(),
  label: z.string(),
  /** Stable backend identifier; label is presentation-only. */
  automationType: z.string().optional(),
  x: z.number(),
  y: z.number(),
});

const legacyAutomationSchema = z.object({
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

const backendAutomationSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  ownerId: z.union([z.string(), z.number()]).transform(Number).optional(),
  name: z.string(),
  triggerType: z.string(),
  triggerConfig: z.union([z.record(z.string(), z.unknown()), z.string()]).optional(),
  actions: z.union([z.array(z.unknown()), z.string()]).optional(),
  isActive: z.boolean().default(false),
  lastRunAt: z.string().nullable().optional(),
  runCount: z.number().catch(0),
  createdAt: z.string(),
  updatedAt: z.string().optional(),
});

function automationTriggerLabel(triggerType: string): string {
  const labels: Record<string, string> = {
    order_created: 'Order created',
    customer_signup: 'New customer',
    loyalty_tier_up: 'Tier upgrade',
    loyalty_tier_down: 'Tier downgrade',
    date_birthday: 'Birthday',
    inactivity: 'Inactive customer',
    tag_added: 'Tag added',
    points_threshold: 'Points threshold',
  };
  return labels[triggerType] ?? triggerType;
}

function automationActionType(action: unknown): string {
  if (typeof action === 'string') return action;
  if (action && typeof action === 'object') {
    const value = action as Record<string, unknown>;
    return String(value.type ?? value.action ?? value.automationType ?? value.label ?? '');
  }
  return '';
}

function actionLabel(action: unknown): string {
  const labels: Record<string, string> = {
    send_email: 'Send email',
    send_sms: 'Send SMS',
    award_points: 'Award points',
    tag_add: 'Add tag',
  };
  const type = automationActionType(action);
  return (labels[type] ?? type) || 'Action';
}

/** Normalize the backend automation payload while retaining the legacy UI fields. */
const automationSchemaDefinition = z
  .union([legacyAutomationSchema, backendAutomationSchema])
  .transform((value) => {
    if ('triggerType' in value) {
      const actions =
        typeof value.actions === 'string'
          ? (() => {
              try {
                return JSON.parse(value.actions) as unknown[];
              } catch {
                return [];
              }
            })()
          : (value.actions ?? []);
      const trigger = automationTriggerLabel(value.triggerType);
      return {
        id: value.id,
        name: value.name,
        description: `${trigger} · ${actions.length} action(s)`,
        trigger,
        triggerType: value.triggerType,
        triggerConfig: value.triggerConfig,
        actions,
        status: value.isActive ? 'active' : 'paused',
        isActive: value.isActive,
        enrolled: value.runCount,
        completed: 0,
        runCount: value.runCount,
        lastRunAt: value.lastRunAt,
        createdAt: value.createdAt,
        updatedAt: value.updatedAt,
        nodes: [
          {
            id: 'n1',
            type: 'trigger',
            label: trigger,
            automationType: value.triggerType,
            x: 0,
            y: 0,
          },
          ...actions.map((action, index) => ({
            id: `n${index + 2}`,
            type: 'action',
            label: actionLabel(action),
            automationType: automationActionType(action),
            x: 0,
            y: (index + 1) * 80,
          })),
        ],
      };
    }
    return value;
  });

export interface Automation {
  id: number;
  name: string;
  description: string;
  trigger: string;
  status: 'active' | 'paused' | 'draft';
  enrolled: number;
  completed: number;
  createdAt: string;
  updatedAt?: string;
  nodes?: z.infer<typeof automationNodeSchema>[];
  triggerType?: string;
  triggerConfig?: Record<string, unknown> | string;
  actions?: unknown[];
  isActive?: boolean;
  runCount?: number;
  lastRunAt?: string | null;
}

export const automationSchema: z.ZodType<Automation> =
  automationSchemaDefinition as z.ZodType<Automation>;

// -- Domain constants ---------------------------------------------------------

export const AUTOMATION_TRIGGERS = [
  'new_customer',
  'birthday',
  'visit',
  'points_threshold',
  'inactivity',
] as const;
export const AUTOMATION_STATUSES = ['active', 'paused', 'draft'] as const;
// -- Type guards --------------------------------------------------------------

export function isAutomationActive(automation: Automation): boolean {
  return automation.status === 'active';
}

export function isAutomationDraft(automation: Automation): boolean {
  return automation.status === 'draft';
}
