import { describe, it, expect } from 'vitest';
import {
  notificationLabel,
  notificationTemplate,
  interpolateTemplate,
  allNotificationEvents,
} from './notification';

describe('notificationLabel', () => {
  it('returns correct labels for known events', () => {
    expect(notificationLabel('application_received')).toBe('Application Received');
    expect(notificationLabel('offer_sent')).toBe('Offer Sent');
  });

  it('returns the event key for unknown events', () => {
    expect(notificationLabel('custom_event')).toBe('custom_event');
  });
});

describe('notificationTemplate', () => {
  it('returns a template string for known events', () => {
    const tmpl = notificationTemplate('application_received');
    expect(tmpl).toContain('{{position}}');
  });

  it('returns empty string for unknown events', () => {
    expect(notificationTemplate('unknown')).toBe('');
  });
});

describe('interpolateTemplate', () => {
  it('replaces placeholders with values', () => {
    const result = interpolateTemplate('Hello {{name}}!', { name: 'Alice' });
    expect(result).toBe('Hello Alice!');
  });

  it('leaves missing placeholders unchanged', () => {
    const result = interpolateTemplate('Hello {{name}}!', {});
    expect(result).toBe('Hello {{name}}!');
  });

  it('replaces multiple different placeholders', () => {
    const result = interpolateTemplate('{{greeting}} {{name}}', {
      greeting: 'Hi',
      name: 'Bob',
    });
    expect(result).toBe('Hi Bob');
  });
});

describe('allNotificationEvents', () => {
  it('returns all six notification events', () => {
    const events = allNotificationEvents();
    expect(events).toHaveLength(6);
    expect(events).toContain('application_received');
    expect(events).toContain('offer_accepted');
  });
});
