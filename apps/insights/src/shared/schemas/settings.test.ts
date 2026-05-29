import { describe, it, expect } from 'vitest';
import { insightsSettingsSchema, scheduledReportSchema } from './settings';

describe('insightsSettingsSchema', () => {
  it('parses a full settings object', () => {
    const data = {
      id: 'set-1',
      defaultPeriod: 'monthly',
      defaultDateRange: 'last_30_days',
      currency: 'USD',
      timezone: 'America/New_York',
      emailReports: true,
      emailRecipients: ['admin@example.com'],
      reportFrequency: 'weekly',
    };
    const result = insightsSettingsSchema.parse(data);
    expect(result.currency).toBe('USD');
    expect(result.emailReports).toBe(true);
    expect(result.emailRecipients).toHaveLength(1);
  });

  it('coerces numeric id to string', () => {
    const result = insightsSettingsSchema.parse({ id: 42 });
    expect(result.id).toBe('42');
  });

  it('accepts a null reportFrequency', () => {
    const result = insightsSettingsSchema.parse({ reportFrequency: null });
    expect(result.reportFrequency).toBeNull();
  });

  it('accepts an empty object (all fields optional)', () => {
    expect(() => insightsSettingsSchema.parse({})).not.toThrow();
  });

  it('accepts an empty emailRecipients array', () => {
    const result = insightsSettingsSchema.parse({ emailRecipients: [] });
    expect(result.emailRecipients).toEqual([]);
  });
});

describe('scheduledReportSchema', () => {
  it('parses a full scheduled report', () => {
    const data = {
      id: 'sr-1',
      name: 'Weekly Summary',
      templateId: 'tpl-001',
      frequency: 'weekly',
      recipients: ['cfo@example.com'],
      enabled: true,
      lastRunAt: '2026-04-01T00:00:00Z',
      nextRunAt: '2026-04-08T00:00:00Z',
    };
    const result = scheduledReportSchema.parse(data);
    expect(result.id).toBe('sr-1');
    expect(result.enabled).toBe(true);
    expect(result.frequency).toBe('weekly');
  });

  it('coerces numeric id to string', () => {
    const result = scheduledReportSchema.parse({ id: 99 });
    expect(result.id).toBe('99');
  });

  it('accepts null lastRunAt and nextRunAt', () => {
    const result = scheduledReportSchema.parse({ id: '1', lastRunAt: null, nextRunAt: null });
    expect(result.lastRunAt).toBeNull();
    expect(result.nextRunAt).toBeNull();
  });

  it('accepts an empty recipients array', () => {
    const result = scheduledReportSchema.parse({ id: '1', recipients: [] });
    expect(result.recipients).toEqual([]);
  });

  it('accepts a data array field', () => {
    const result = scheduledReportSchema.parse({ id: '1', data: [{ row: 1 }] });
    expect(Array.isArray(result.data)).toBe(true);
  });
});
