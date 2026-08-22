import { describe, expect, it } from 'vitest';
import {
  actionTypeFor,
  mergeAutomationData,
  toBackendRequest,
  triggerTypeFor,
} from './useAutomations';

describe('automation wire mappings', () => {
  it('maps every visible trigger label to an Onion trigger enum', () => {
    expect(triggerTypeFor('Order created')).toBe('order_created');
    expect(triggerTypeFor('New customer')).toBe('customer_signup');
    expect(triggerTypeFor('Tier upgrade')).toBe('loyalty_tier_up');
    expect(triggerTypeFor('Tag added')).toBe('tag_added');
    expect(triggerTypeFor('Points threshold')).toBe('points_threshold');
  });

  it('maps visible action labels to backend action enums', () => {
    expect(actionTypeFor('Send email')).toBe('send_email');
    expect(actionTypeFor('Send SMS')).toBe('send_sms');
    expect(actionTypeFor('Award points')).toBe('award_points');
  });

  it('preserves required fields when an edit supplies only a changed field', () => {
    const current = {
      id: 7,
      name: 'Welcome series',
      description: 'Customer signup · 1 action(s)',
      trigger: 'New customer',
      triggerType: 'customer_signup',
      triggerConfig: { days: 0 },
      actions: [{ type: 'send_email', body: 'Welcome' }],
      status: 'paused' as const,
      isActive: false,
      enrolled: 0,
      completed: 0,
      createdAt: '2026-08-21T00:00:00Z',
    };

    expect(toBackendRequest(mergeAutomationData(current, { isActive: true }))).toEqual({
      name: 'Welcome series',
      triggerType: 'customer_signup',
      triggerConfig: { days: 0 },
      actions: [{ type: 'send_email', body: 'Welcome' }],
      isActive: true,
    });
  });
});
