import { describe, expect, it } from 'vitest';
import {
  groupAndCount,
  percentageChange,
  periodStartDate,
  reportFilename,
  reportFormatLabel,
  reportFormatMimeType,
  reportPeriodLabel,
} from './reportUtils';

describe('reportPeriodLabel', () => {
  it('returns correct label for each period', () => {
    expect(reportPeriodLabel('weekly')).toBe('Weekly');
    expect(reportPeriodLabel('monthly')).toBe('Monthly');
    expect(reportPeriodLabel('quarterly')).toBe('Quarterly');
    expect(reportPeriodLabel('yearly')).toBe('Yearly');
  });
});

describe('reportFormatLabel', () => {
  it('returns correct label for each format', () => {
    expect(reportFormatLabel('csv')).toBe('CSV');
    expect(reportFormatLabel('pdf')).toBe('PDF');
    expect(reportFormatLabel('xlsx')).toBe('Excel');
  });
});

describe('reportFormatMimeType', () => {
  it('returns correct MIME type', () => {
    expect(reportFormatMimeType('csv')).toBe('text/csv');
    expect(reportFormatMimeType('pdf')).toBe('application/pdf');
    expect(reportFormatMimeType('xlsx')).toContain('spreadsheetml');
  });
});

describe('periodStartDate', () => {
  it('returns start of month for monthly', () => {
    const result = periodStartDate('2024-03-15', 'monthly');
    expect(result).toBe('2024-03-01');
  });

  it('returns start of year for yearly', () => {
    const result = periodStartDate('2024-06-20', 'yearly');
    expect(result).toBe('2024-01-01');
  });

  it('returns start of quarter for quarterly', () => {
    const result = periodStartDate('2024-05-10', 'quarterly');
    expect(result).toBe('2024-04-01');
  });
});

describe('percentageChange', () => {
  it('calculates percentage change correctly', () => {
    expect(percentageChange(100, 150)).toBeCloseTo(50);
    expect(percentageChange(200, 100)).toBeCloseTo(-50);
    expect(percentageChange(100, 100)).toBeCloseTo(0);
  });

  it('returns null when previous is 0', () => {
    expect(percentageChange(0, 100)).toBeNull();
  });
});

describe('reportFilename', () => {
  it('generates correct filename', () => {
    const result = reportFilename('Employee Report', 'monthly', '2024-01-01', 'pdf');
    expect(result).toBe('employee_report_monthly_2024-01-01.pdf');
  });

  it('handles different formats', () => {
    const result = reportFilename('Sales', 'yearly', '2024-01-01', 'xlsx');
    expect(result).toBe('sales_yearly_2024-01-01.xlsx');
  });
});

describe('groupAndCount', () => {
  it('groups and counts records by key', () => {
    const records = [
      { dept: 'Engineering' },
      { dept: 'HR' },
      { dept: 'Engineering' },
      { dept: 'Sales' },
    ];
    const result = groupAndCount(records, (r) => r.dept);
    expect(result['Engineering']).toBe(2);
    expect(result['HR']).toBe(1);
    expect(result['Sales']).toBe(1);
  });

  it('returns empty object for empty array', () => {
    expect(groupAndCount([], (r: { k: string }) => r.k)).toEqual({});
  });
});
