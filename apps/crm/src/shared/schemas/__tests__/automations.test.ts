import { describe, it, expect } from 'vitest';
import {
  automationSchema,
  AUTOMATION_TRIGGERS,
  AUTOMATION_STATUSES,
  isAutomationActive,
  isAutomationDraft,
} from '../automations';

const baseAutomation = {
  id: 1,
  name: 'Birthday Reward',
  description: 'Send a reward on customer birthday',
  trigger: 'birthday',
  status: 'active',
  enrolled: 120,
  completed: 85,
  createdAt: '2024-01-10',
};

describe('automationSchema', () => {
  it('parses a valid automation object', () => {
    const result = automationSchema.parse(baseAutomation);
    expect(result.id).toBe(1);
    expect(result.name).toBe('Birthday Reward');
    expect(result.status).toBe('active');
  });

  it('coerces a string id to number', () => {
    const result = automationSchema.parse({ ...baseAutomation, id: '7' });
    expect(result.id).toBe(7);
  });

  it('falls back to draft for an unknown status', () => {
    const result = automationSchema.parse({ ...baseAutomation, status: 'unknown' });
    expect(result.status).toBe('draft');
  });

  it('accepts optional nodes array', () => {
    const result = automationSchema.parse({
      ...baseAutomation,
      nodes: [{ id: 'n1', type: 'trigger', label: 'Start', x: 0, y: 0 }],
    });
    expect(result.nodes).toHaveLength(1);
  });

  it('omits nodes when not provided', () => {
    const result = automationSchema.parse(baseAutomation);
    expect(result.nodes).toBeUndefined();
  });
});

describe('AUTOMATION_TRIGGERS', () => {
  it('contains all expected trigger types', () => {
    expect(AUTOMATION_TRIGGERS).toContain('new_customer');
    expect(AUTOMATION_TRIGGERS).toContain('birthday');
    expect(AUTOMATION_TRIGGERS).toContain('visit');
    expect(AUTOMATION_TRIGGERS).toContain('points_threshold');
    expect(AUTOMATION_TRIGGERS).toContain('inactivity');
  });

  it('has exactly 5 triggers', () => {
    expect(AUTOMATION_TRIGGERS).toHaveLength(5);
  });
});

describe('AUTOMATION_STATUSES', () => {
  it('contains active, paused, and draft', () => {
    expect(AUTOMATION_STATUSES).toContain('active');
    expect(AUTOMATION_STATUSES).toContain('paused');
    expect(AUTOMATION_STATUSES).toContain('draft');
  });
});

describe('isAutomationActive', () => {
  it('returns true for active automations', () => {
    const automation = automationSchema.parse({ ...baseAutomation, status: 'active' });
    expect(isAutomationActive(automation)).toBe(true);
  });

  it('returns false for non-active automations', () => {
    const automation = automationSchema.parse({ ...baseAutomation, status: 'paused' });
    expect(isAutomationActive(automation)).toBe(false);
  });
});

describe('isAutomationDraft', () => {
  it('returns true for draft automations', () => {
    const automation = automationSchema.parse({ ...baseAutomation, status: 'draft' });
    expect(isAutomationDraft(automation)).toBe(true);
  });

  it('returns false for non-draft automations', () => {
    const automation = automationSchema.parse({ ...baseAutomation, status: 'active' });
    expect(isAutomationDraft(automation)).toBe(false);
  });
});
