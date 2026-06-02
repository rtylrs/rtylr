import { describe, it, expect } from 'vitest';
import { reportTemplateSchema } from './reports';

describe('reportTemplateSchema', () => {
  it('parses a report template matching the backend contract', () => {
    const data = {
      id: 'revenue-monthly',
      name: 'Monthly Revenue',
      description: 'Month-by-month paid revenue over the last 12 months.',
    };
    const result = reportTemplateSchema.parse(data);
    expect(result.id).toBe('revenue-monthly');
    expect(result.name).toBe('Monthly Revenue');
    expect(result.description).toBe('Month-by-month paid revenue over the last 12 months.');
  });

  it('rejects a missing description (backend always sends it)', () => {
    expect(() => reportTemplateSchema.parse({ id: '1', name: 'Sales' })).toThrow();
  });

  it('rejects a null description', () => {
    expect(() =>
      reportTemplateSchema.parse({ id: '1', name: 'Sales', description: null })
    ).toThrow();
  });

  it('rejects a missing id', () => {
    expect(() => reportTemplateSchema.parse({ name: 'Sales', description: 'x' })).toThrow();
  });

  it('rejects a missing name', () => {
    expect(() => reportTemplateSchema.parse({ id: '1', description: 'x' })).toThrow();
  });

  it('ignores fields the backend does not send', () => {
    const result = reportTemplateSchema.parse({
      id: '1',
      name: 'Sales',
      description: 'desc',
      category: 'Revenue',
      params: [{ key: 'q', label: 'Query' }],
      data: [{ date: '2026-01', value: 100 }],
    });
    expect(result).toEqual({ id: '1', name: 'Sales', description: 'desc' });
  });
});
