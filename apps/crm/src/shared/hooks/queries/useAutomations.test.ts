import { describe, expect, it } from 'vitest';
import { actionTypeFor, triggerTypeFor } from './useAutomations';

describe('automation wire mappings', () => {
  it('maps every visible trigger label to an Onion trigger enum', () => {
    expect(triggerTypeFor('Order created')).toBe('order_created');
    expect(triggerTypeFor('New customer')).toBe('customer_signup');
    expect(triggerTypeFor('Tier upgrade')).toBe('loyalty_tier_up');
  });

  it('maps visible action labels to backend action enums', () => {
    expect(actionTypeFor('Send email')).toBe('send_email');
    expect(actionTypeFor('Send SMS')).toBe('send_sms');
    expect(actionTypeFor('Award points')).toBe('award_points');
  });
});
