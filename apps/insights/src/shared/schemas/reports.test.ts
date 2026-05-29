import { describe, it, expect } from 'vitest';
import { reportTemplateSchema } from './reports';

describe('reportTemplateSchema', () => {
  it('parses a full report template', () => {
    const data = {
      id: 'rpt-001',
      name: 'Monthly Revenue',
      description: 'Revenue breakdown by month',
      category: 'Revenue',
      params: [
        { key: 'start_date', label: 'Start Date', type: 'date', required: true },
        { key: 'end_date', label: 'End Date', type: 'date', required: true },
      ],
    };
    const result = reportTemplateSchema.parse(data);
    expect(result.id).toBe('rpt-001');
    expect(result.name).toBe('Monthly Revenue');
    expect(result.params).toHaveLength(2);
    expect(result.params?.[0].required).toBe(true);
  });

  it('coerces numeric id to string', () => {
    const result = reportTemplateSchema.parse({ id: 7, name: 'Daily Orders' });
    expect(result.id).toBe('7');
  });

  it('accepts a null description', () => {
    const result = reportTemplateSchema.parse({ id: '1', name: 'Sales', description: null });
    expect(result.description).toBeNull();
  });

  it('accepts a null category', () => {
    const result = reportTemplateSchema.parse({ id: '1', name: 'Sales', category: null });
    expect(result.category).toBeNull();
  });

  it('accepts an empty params array', () => {
    const result = reportTemplateSchema.parse({ id: '1', name: 'Empty', params: [] });
    expect(result.params).toEqual([]);
  });

  it('accepts a data array of unknown shape', () => {
    const result = reportTemplateSchema.parse({
      id: '1',
      name: 'Data Report',
      data: [{ date: '2026-01', value: 100 }],
    });
    expect(Array.isArray(result.data)).toBe(true);
    expect(result.data).toHaveLength(1);
  });

  it('accepts params with optional type and required fields missing', () => {
    const result = reportTemplateSchema.parse({
      id: '1',
      name: 'Simple',
      params: [{ key: 'q', label: 'Query' }],
    });
    expect(result.params?.[0].type).toBeUndefined();
    expect(result.params?.[0].required).toBeUndefined();
  });
});
