import { describe, it, expect } from 'vitest';
import { csvEscape, toCsv, parseCsv, csvFilename } from './csvUtils';

describe('csvEscape', () => {
  it('returns plain values unchanged', () => {
    expect(csvEscape('hello')).toBe('hello');
  });

  it('wraps values containing commas in quotes', () => {
    expect(csvEscape('hello, world')).toBe('"hello, world"');
  });

  it('escapes internal double quotes', () => {
    expect(csvEscape('say "hi"')).toBe('"say ""hi"""');
  });

  it('wraps values containing newlines', () => {
    expect(csvEscape('line1\nline2')).toBe('"line1\nline2"');
  });

  it('returns empty string for null', () => {
    expect(csvEscape(null)).toBe('');
  });

  it('converts numbers to strings', () => {
    expect(csvEscape(42)).toBe('42');
  });
});

describe('toCsv', () => {
  const rows = [
    { name: 'Alice', age: 30 },
    { name: 'Bob', age: 25 },
  ];

  it('produces a header row', () => {
    const csv = toCsv(rows);
    expect(csv.split('\n')[0]).toBe('name,age');
  });

  it('produces data rows', () => {
    const csv = toCsv(rows);
    const lines = csv.split('\n');
    expect(lines[1]).toBe('Alice,30');
    expect(lines[2]).toBe('Bob,25');
  });

  it('respects custom column order', () => {
    const csv = toCsv(rows, ['age', 'name']);
    expect(csv.split('\n')[0]).toBe('age,name');
  });

  it('returns empty string for empty rows', () => {
    expect(toCsv([])).toBe('');
  });
});

describe('parseCsv', () => {
  it('parses a simple CSV', () => {
    const result = parseCsv('name,age\nAlice,30\nBob,25');
    expect(result).toHaveLength(3);
    expect(result[0]).toEqual(['name', 'age']);
    expect(result[1]).toEqual(['Alice', '30']);
  });

  it('handles quoted fields', () => {
    const result = parseCsv('"hello, world",test');
    expect(result[0]![0]).toBe('hello, world');
  });

  it('ignores empty lines', () => {
    const result = parseCsv('a,b\n\nc,d\n');
    expect(result).toHaveLength(2);
  });
});

describe('csvFilename', () => {
  it('includes base name and date', () => {
    const result = csvFilename('report', new Date('2025-06-15'));
    expect(result).toBe('report_2025-06-15.csv');
  });

  it('appends .csv extension', () => {
    expect(csvFilename('export', new Date('2025-01-01'))).toMatch(/\.csv$/);
  });
});
