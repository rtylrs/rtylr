import { toIsoDate } from './dateUtils';
/**
 * Data export helpers for the ERP app (CSV and JSON export formatting).
 * All functions are pure with no side-effects.
 */

export interface ExportColumn<T> {
  key: keyof T;
  header: string;
  format?: (value: T[keyof T]) => string;
}

/** Escape a CSV cell value (quote if it contains commas, quotes, or newlines). */
export function escapeCsvCell(value: unknown): string {
  const str = String(value ?? '');
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

/** Build a CSV row from an array of values. */
export function buildCsvRow(values: unknown[]): string {
  return values.map(escapeCsvCell).join(',');
}

/** Build a complete CSV string from rows and column definitions. */
export function buildCsv<T>(rows: T[], columns: ExportColumn<T>[]): string {
  const header = buildCsvRow(columns.map((c) => c.header));
  const dataRows = rows.map((row) =>
    buildCsvRow(
      columns.map((col) => {
        const value = row[col.key];
        return col.format ? col.format(value) : value;
      })
    )
  );
  return [header, ...dataRows].join('\n');
}

/** Convert an object to a flat key=value string (for simple exports). */
export function objectToKeyValue(obj: Record<string, unknown>, separator = ': '): string {
  return Object.entries(obj)
    .map(([key, value]) => `${key}${separator}${value}`)
    .join('\n');
}

/** Sanitize a filename by replacing disallowed characters. */
export function sanitizeFilename(name: string): string {
  return name
    .replace(/[/\\:*?"<>|]/g, '-')
    .replace(/\s+/g, '_')
    .replace(/-+/g, '-')
    .slice(0, 100);
}

/** Build an export filename with date suffix. */
export function exportFilename(prefix: string, extension: string, date?: Date): string {
  const d = date ?? new Date();
  const dateStr = toIsoDate(d);
  return `${sanitizeFilename(prefix)}_${dateStr}.${extension}`;
}

/** Compute file size description for a string. */
export function estimateFileSizeLabel(content: string): string {
  const bytes = new TextEncoder().encode(content).length;
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** Convert an array of objects to JSON string. */
export function toJsonExport(data: unknown[], pretty = false): string {
  return JSON.stringify(data, null, pretty ? 2 : undefined);
}
