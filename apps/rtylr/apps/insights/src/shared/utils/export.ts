// CSV export helpers for the insights feature.

/**
 * rowsToCsv converts an array of objects to a CSV string.
 * All keys from the first row are used as headers.
 */
export function rowsToCsv<T extends Record<string, unknown>>(rows: T[]): string {
  if (rows.length === 0) return '';
  const headers = Object.keys(rows[0]);
  const escape = (val: unknown): string => {
    const s = String(val ?? '');
    if (s.includes(',') || s.includes('"') || s.includes('\n')) {
      return `"${s.replace(/"/g, '""')}"`;
    }
    return s;
  };
  const lines = [
    headers.join(','),
    ...rows.map((row) => headers.map((h) => escape(row[h])).join(',')),
  ];
  return lines.join('\n');
}

/**
 * downloadCsv triggers a browser download for the given CSV string.
 * No-ops in non-browser environments.
 */
export function downloadCsv(csv: string, filename: string): void {
  if (typeof document === 'undefined') return;
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/**
 * sanitizeCsvFilename removes characters unsafe for filenames and trims whitespace.
 */
export function sanitizeCsvFilename(name: string): string {
  return name.replace(/[^a-zA-Z0-9_\-. ]/g, '').trim();
}
