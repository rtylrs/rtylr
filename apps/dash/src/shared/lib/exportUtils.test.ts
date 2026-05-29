import { describe, it, expect } from 'vitest';
import {
  filenameTimestamp,
  buildFilename,
  mimeType,
  toJsonExport,
  estimateBlobSize,
  supportsColumnSelection,
  clampExportRows,
} from './exportUtils';

describe('filenameTimestamp', () => {
  it('produces a string with no colons or dots', () => {
    const ts = filenameTimestamp(new Date('2026-05-01T12:30:00Z'));
    expect(ts).not.toContain(':');
    expect(ts).not.toContain('.');
  });

  it('produces a 19-character string', () => {
    const ts = filenameTimestamp(new Date('2026-05-01T12:30:00Z'));
    expect(ts).toHaveLength(19);
  });

  it('accepts an ISO string', () => {
    const ts = filenameTimestamp('2026-01-01T00:00:00Z');
    expect(ts).toBe('2026-01-01T00-00-00');
  });
});

describe('buildFilename', () => {
  it('combines base, timestamp, and extension', () => {
    const filename = buildFilename('orders', 'csv', new Date('2026-05-01T12:00:00Z'));
    expect(filename).toMatch(/^orders_2026-05-01T12-00-00\.csv$/);
  });

  it('adds a leading dot to extension when missing', () => {
    const filename = buildFilename('report', 'pdf', new Date('2026-01-01T00:00:00Z'));
    expect(filename.endsWith('.pdf')).toBe(true);
  });

  it('keeps a leading dot when already present', () => {
    const filename = buildFilename('data', '.json', new Date('2026-01-01T00:00:00Z'));
    expect(filename.endsWith('.json')).toBe(true);
    expect(filename).not.toContain('..json');
  });
});

describe('mimeType', () => {
  it('returns correct MIME for csv', () => {
    expect(mimeType('csv')).toContain('text/csv');
  });

  it('returns correct MIME for json', () => {
    expect(mimeType('json')).toBe('application/json');
  });

  it('returns correct MIME for pdf', () => {
    expect(mimeType('pdf')).toBe('application/pdf');
  });

  it('returns correct MIME for xlsx', () => {
    expect(mimeType('xlsx')).toContain('spreadsheetml');
  });
});

describe('toJsonExport', () => {
  it('serializes data to JSON', () => {
    const json = toJsonExport({ id: 1, name: 'Alice' });
    expect(JSON.parse(json)).toEqual({ id: 1, name: 'Alice' });
  });

  it('indents with 2 spaces by default', () => {
    const json = toJsonExport({ a: 1 });
    expect(json).toContain('\n  ');
  });
});

describe('estimateBlobSize', () => {
  it('returns bytes for small strings', () => {
    expect(estimateBlobSize('hello')).toBe('5 B');
  });

  it('returns KB for strings over 1024 bytes', () => {
    const large = 'x'.repeat(2048);
    expect(estimateBlobSize(large)).toContain('KB');
  });
});

describe('supportsColumnSelection', () => {
  it('returns true for csv', () => {
    expect(supportsColumnSelection('csv')).toBe(true);
  });

  it('returns true for xlsx', () => {
    expect(supportsColumnSelection('xlsx')).toBe(true);
  });

  it('returns false for pdf', () => {
    expect(supportsColumnSelection('pdf')).toBe(false);
  });

  it('returns false for json', () => {
    expect(supportsColumnSelection('json')).toBe(false);
  });
});

describe('clampExportRows', () => {
  it('returns all rows when under the limit', () => {
    const result = clampExportRows([1, 2, 3], 10);
    expect(result.rows).toHaveLength(3);
    expect(result.truncated).toBe(false);
  });

  it('clamps rows and sets truncated to true', () => {
    const result = clampExportRows([1, 2, 3, 4, 5], 3);
    expect(result.rows).toHaveLength(3);
    expect(result.truncated).toBe(true);
  });

  it('is not truncated when exactly at the limit', () => {
    const result = clampExportRows([1, 2, 3], 3);
    expect(result.truncated).toBe(false);
  });
});
