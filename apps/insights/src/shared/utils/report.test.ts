import { describe, it, expect } from 'vitest';
import {
  activeScheduledReports,
  reportsByTemplate,
  cadenceLabel,
  reportChannelLabel,
  findTemplate,
  sortScheduledReports,
} from './report';
import type { ReportTemplate, ScheduledReport } from '../hooks/queries/types';

function makeReport(overrides: Partial<ScheduledReport> = {}): ScheduledReport {
  return {
    id: 1,
    templateId: 'revenue',
    cadence: 'weekly',
    channel: 'email',
    recipient: 'admin@example.com',
    isActive: true,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('activeScheduledReports', () => {
  it('returns only active reports', () => {
    const reports = [makeReport({ id: 1, isActive: true }), makeReport({ id: 2, isActive: false })];
    expect(activeScheduledReports(reports)).toHaveLength(1);
  });
});

describe('reportsByTemplate', () => {
  it('groups by templateId', () => {
    const reports = [
      makeReport({ templateId: 'rev', id: 1 }),
      makeReport({ templateId: 'rev', id: 2 }),
      makeReport({ templateId: 'ops', id: 3 }),
    ];
    const groups = reportsByTemplate(reports);
    expect(groups.get('rev')).toHaveLength(2);
  });
});

describe('cadenceLabel', () => {
  it('returns human-readable cadence', () => {
    expect(cadenceLabel('daily')).toBe('Daily');
    expect(cadenceLabel('monthly')).toBe('Monthly');
  });

  it('returns cadence itself for unknown values', () => {
    expect(cadenceLabel('custom')).toBe('custom');
  });
});

describe('reportChannelLabel', () => {
  it('returns human-readable channel', () => {
    expect(reportChannelLabel('email')).toBe('Email');
  });
});

describe('findTemplate', () => {
  it('finds template by id', () => {
    const templates: ReportTemplate[] = [
      { id: 'a', name: 'A', description: '' },
      { id: 'b', name: 'B', description: '' },
    ];
    expect(findTemplate(templates, 'b')?.id).toBe('b');
  });

  it('returns undefined when not found', () => {
    expect(findTemplate([], 'x')).toBeUndefined();
  });
});

describe('sortScheduledReports', () => {
  it('sorts newest first', () => {
    const reports = [
      makeReport({ id: 1, createdAt: '2026-01-01T00:00:00Z' }),
      makeReport({ id: 2, createdAt: '2026-06-01T00:00:00Z' }),
    ];
    expect(sortScheduledReports(reports)[0].id).toBe(2);
  });
});
