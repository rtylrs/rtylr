import { describe, it, expect } from 'vitest';
import { escapeCsvField, toCsvString } from './csvExport';

describe('escapeCsvField', () => {
  it('returns the value as a string for plain text', () => {
    expect(escapeCsvField('hello')).toBe('hello');
  });

  it('wraps values containing commas in double quotes', () => {
    expect(escapeCsvField('hello, world')).toBe('"hello, world"');
  });

  it('wraps values containing double quotes and escapes them', () => {
    expect(escapeCsvField('say "hi"')).toBe('"say ""hi"""');
  });

  it('wraps values containing newlines in double quotes', () => {
    expect(escapeCsvField('line1\nline2')).toBe('"line1\nline2"');
  });

  it('converts numbers to strings', () => {
    expect(escapeCsvField(42)).toBe('42');
  });

  it('converts null to an empty string', () => {
    expect(escapeCsvField(null)).toBe('');
  });

  it('converts undefined to an empty string', () => {
    expect(escapeCsvField(undefined)).toBe('');
  });

  it('converts booleans to strings', () => {
    expect(escapeCsvField(true)).toBe('true');
    expect(escapeCsvField(false)).toBe('false');
  });
});

describe('toCsvString', () => {
  it('returns an empty string for an empty array', () => {
    expect(toCsvString([])).toBe('');
  });

  it('produces a header row and a data row', () => {
    const rows = [{ name: 'Widget', price: 10 }];
    const csv = toCsvString(rows);
    const lines = csv.split('\n');
    expect(lines[0]).toBe('name,price');
    expect(lines[1]).toBe('Widget,10');
  });

  it('handles multiple rows', () => {
    const rows = [
      { a: 1, b: 2 },
      { a: 3, b: 4 },
    ];
    const lines = toCsvString(rows).split('\n');
    expect(lines).toHaveLength(3);
    expect(lines[2]).toBe('3,4');
  });

  it('escapes fields with commas', () => {
    const rows = [{ name: 'Bolt, Small', qty: 5 }];
    const csv = toCsvString(rows);
    expect(csv).toContain('"Bolt, Small"');
  });

  it('treats null values as empty strings', () => {
    const rows = [{ name: null, qty: 5 }];
    const csv = toCsvString(rows);
    const dataLine = csv.split('\n')[1];
    expect(dataLine).toBe(',5');
  });
});
