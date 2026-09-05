import { describe, it, expect } from 'vitest';
import {
  escapeCsvCell,
  buildCsvRow,
  buildCsv,
  objectToKeyValue,
  sanitizeFilename,
  exportFilename,
  toJsonExport,
} from './exportUtils';

describe('escapeCsvCell', () => {
  it('returns plain string unchanged', () => {
    expect(escapeCsvCell('hello')).toBe('hello');
  });

  it('quotes string containing comma', () => {
    expect(escapeCsvCell('a,b')).toBe('"a,b"');
  });

  it('escapes internal double quotes', () => {
    expect(escapeCsvCell('say "hello"')).toBe('"say ""hello"""');
  });

  it('converts null and undefined to empty string', () => {
    expect(escapeCsvCell(null)).toBe('');
    expect(escapeCsvCell(undefined)).toBe('');
  });
});

describe('buildCsvRow', () => {
  it('joins values with commas', () => {
    expect(buildCsvRow(['a', 'b', 'c'])).toBe('a,b,c');
  });

  it('quotes values with commas', () => {
    expect(buildCsvRow(['a,b', 'c'])).toBe('"a,b",c');
  });
});

describe('buildCsv', () => {
  it('builds CSV with header and data rows', () => {
    const rows = [
      { id: 1, name: 'Widget', price: 9.99 },
      { id: 2, name: 'Gadget', price: 19.99 },
    ];
    const columns = [
      { key: 'id' as const, header: 'ID' },
      { key: 'name' as const, header: 'Name' },
      { key: 'price' as const, header: 'Price', format: (v: number) => `$${v}` },
    ];
    const csv = buildCsv(rows, columns);
    expect(csv).toContain('ID,Name,Price');
    expect(csv).toContain('1,Widget,$9.99');
    expect(csv).toContain('2,Gadget,$19.99');
  });
});

describe('objectToKeyValue', () => {
  it('converts object to key: value format', () => {
    const result = objectToKeyValue({ name: 'Widget', price: 9.99 });
    expect(result).toContain('name: Widget');
    expect(result).toContain('price: 9.99');
  });
});

describe('sanitizeFilename', () => {
  it('replaces disallowed characters', () => {
    expect(sanitizeFilename('file/name:test')).toBe('file-name-test');
  });

  it('replaces spaces with underscores', () => {
    expect(sanitizeFilename('my file name')).toBe('my_file_name');
  });

  it('truncates to 100 characters', () => {
    const long = 'a'.repeat(150);
    expect(sanitizeFilename(long)).toHaveLength(100);
  });
});

describe('exportFilename', () => {
  it('builds filename with prefix, date, and extension', () => {
    const date = new Date(2026, 4, 29);
    const filename = exportFilename('inventory', 'csv', date);
    expect(filename).toBe('inventory_2026-05-29.csv');
  });
});

describe('toJsonExport', () => {
  it('serializes array to JSON', () => {
    const result = toJsonExport([{ id: 1 }]);
    expect(JSON.parse(result)).toEqual([{ id: 1 }]);
  });

  it('pretty-prints when requested', () => {
    const result = toJsonExport([{ id: 1 }], true);
    expect(result).toContain('\n');
  });
});
