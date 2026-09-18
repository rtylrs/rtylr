import { describe, it, expect } from 'vitest';
import { rowsToCsv, sanitizeCsvFilename } from './export';

describe('rowsToCsv', () => {
  it('returns empty string for empty rows', () => {
    expect(rowsToCsv([])).toBe('');
  });

  it('generates correct headers and values', () => {
    const csv = rowsToCsv([{ name: 'Alice', age: 30 }]);
    expect(csv).toBe('name,age\nAlice,30');
  });

  it('escapes commas in values', () => {
    const csv = rowsToCsv([{ label: 'a,b' }]);
    expect(csv).toContain('"a,b"');
  });

  it('escapes double quotes in values', () => {
    const csv = rowsToCsv([{ label: 'say "hello"' }]);
    expect(csv).toContain('"say ""hello"""');
  });

  it('handles multiple rows', () => {
    const csv = rowsToCsv([{ x: 1 }, { x: 2 }]);
    const lines = csv.split('\n');
    expect(lines).toHaveLength(3);
  });
});

describe('sanitizeCsvFilename', () => {
  it('removes unsafe characters', () => {
    expect(sanitizeCsvFilename('report/2026.csv')).toBe('report2026.csv');
  });

  it('trims surrounding whitespace', () => {
    expect(sanitizeCsvFilename('  report  ')).toBe('report');
  });

  it('preserves safe characters', () => {
    expect(sanitizeCsvFilename('my-report_2026.csv')).toBe('my-report_2026.csv');
  });
});
