import { describe, it, expect } from 'vitest';
import {
  exportFormatLabel,
  exportFormatMimeType,
  exportScopeLabel,
  generateExportFilename,
  objectsToCSV,
  estimateExportSizeKB,
} from './hrExportUtils';

describe('exportFormatLabel', () => {
  it('returns CSV for csv', () => {
    expect(exportFormatLabel('csv')).toBe('CSV');
  });
  it('returns Excel for xlsx', () => {
    expect(exportFormatLabel('xlsx')).toBe('Excel');
  });
  it('returns JSON for json', () => {
    expect(exportFormatLabel('json')).toBe('JSON');
  });
  it('returns PDF for pdf', () => {
    expect(exportFormatLabel('pdf')).toBe('PDF');
  });
});

describe('exportFormatMimeType', () => {
  it('returns text/csv for csv', () => {
    expect(exportFormatMimeType('csv')).toBe('text/csv');
  });
  it('returns application/json for json', () => {
    expect(exportFormatMimeType('json')).toBe('application/json');
  });
  it('returns application/pdf for pdf', () => {
    expect(exportFormatMimeType('pdf')).toBe('application/pdf');
  });
});

describe('exportScopeLabel', () => {
  it('returns All Records for all', () => {
    expect(exportScopeLabel('all')).toBe('All Records');
  });
  it('returns Selected Records for selected', () => {
    expect(exportScopeLabel('selected')).toBe('Selected Records');
  });
  it('returns Filtered Records for filtered', () => {
    expect(exportScopeLabel('filtered')).toBe('Filtered Records');
  });
});

describe('generateExportFilename', () => {
  it('generates filename with provided date', () => {
    const name = generateExportFilename('Employees', 'csv', '2024-06-01');
    expect(name).toBe('employees_export_2024-06-01.csv');
  });
  it('replaces spaces with underscores', () => {
    const name = generateExportFilename('HR Report', 'xlsx', '2024-01-01');
    expect(name).toBe('hr_report_export_2024-01-01.xlsx');
  });
  it('uses today when no date provided', () => {
    const name = generateExportFilename('test', 'json');
    expect(name).toMatch(/^test_export_\d{4}-\d{2}-\d{2}\.json$/);
  });
});

describe('objectsToCSV', () => {
  it('returns empty string for empty array', () => {
    expect(objectsToCSV([], ['name'])).toBe('');
  });
  it('generates header and rows', () => {
    const data = [
      { name: 'Alice', age: 30 },
      { name: 'Bob', age: 25 },
    ];
    const csv = objectsToCSV(data, ['name', 'age']);
    expect(csv).toBe('name,age\nAlice,30\nBob,25');
  });
  it('wraps values with commas in quotes', () => {
    const data = [{ name: 'Doe, Jane', age: 28 }];
    const csv = objectsToCSV(data, ['name', 'age']);
    expect(csv).toContain('"Doe, Jane"');
  });
});

describe('estimateExportSizeKB', () => {
  it('returns positive number', () => {
    expect(estimateExportSizeKB(100, 10)).toBeGreaterThan(0);
  });
  it('scales with record and field count', () => {
    const small = estimateExportSizeKB(10, 5);
    const large = estimateExportSizeKB(1000, 5);
    expect(large).toBeGreaterThan(small);
  });
});
