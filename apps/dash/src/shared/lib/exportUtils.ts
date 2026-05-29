/**
 * Export utility helpers for the Dash app.
 * All functions are pure with no side-effects unless noted.
 */

/** Convert a JS Date or ISO string to a compact filename-safe timestamp. */
export function filenameTimestamp(date: Date | string = new Date()): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return d.toISOString().replace(/[:.]/g, '-').slice(0, 19);
}

/** Build a default export filename from a base name and optional date. */
export function buildFilename(base: string, extension: string, date?: Date | string): string {
  const ts = filenameTimestamp(date);
  const ext = extension.startsWith('.') ? extension : `.${extension}`;
  return `${base}_${ts}${ext}`;
}

/** Return the MIME type for common export formats. */
export function mimeType(format: 'csv' | 'json' | 'pdf' | 'xlsx'): string {
  const map: Record<string, string> = {
    csv: 'text/csv;charset=utf-8;',
    json: 'application/json',
    pdf: 'application/pdf',
    xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  };
  return map[format] ?? 'application/octet-stream';
}

/** Serialize an object to a pretty-printed JSON string. */
export function toJsonExport(data: unknown, indent = 2): string {
  return JSON.stringify(data, null, indent);
}

/** Estimate the human-readable size of a string blob. */
export function estimateBlobSize(content: string): string {
  const bytes = new TextEncoder().encode(content).length;
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** Return true when the export format supports column selection. */
export function supportsColumnSelection(format: string): boolean {
  return format === 'csv' || format === 'xlsx';
}

/** Clamp the number of rows in an export to a maximum limit. */
export function clampExportRows<T>(rows: T[], limit: number): { rows: T[]; truncated: boolean } {
  if (rows.length <= limit) return { rows, truncated: false };
  return { rows: rows.slice(0, limit), truncated: true };
}
