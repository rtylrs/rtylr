/**
 * CSV export helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Convert a value to a CSV-safe string (escapes quotes and wraps in quotes if needed). */
export function csvEscape(value: unknown): string {
  const str = value == null ? '' : String(value);
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

/** Serialize an array of objects to a CSV string. */
export function toCsv<T extends Record<string, unknown>>(rows: T[], columns?: (keyof T)[]): string {
  if (rows.length === 0) return '';
  const keys = columns ?? (Object.keys(rows[0]!) as (keyof T)[]);
  const header = keys.map((k) => csvEscape(String(k))).join(',');
  const body = rows.map((row) => keys.map((k) => csvEscape(row[k])).join(',')).join('\n');
  return `${header}\n${body}`;
}

/** Parse a simple CSV string into an array of string arrays. */
export function parseCsv(csv: string): string[][] {
  return csv
    .split('\n')
    .filter((line) => line.trim())
    .map((line) => {
      const cells: string[] = [];
      let current = '';
      let inQuotes = false;
      for (let i = 0; i < line.length; i++) {
        const ch = line[i]!;
        if (ch === '"') {
          if (inQuotes && line[i + 1] === '"') {
            current += '"';
            i++;
          } else {
            inQuotes = !inQuotes;
          }
        } else if (ch === ',' && !inQuotes) {
          cells.push(current);
          current = '';
        } else {
          current += ch;
        }
      }
      cells.push(current);
      return cells;
    });
}

/** Generate a filename with timestamp for CSV downloads. */
export function csvFilename(baseName: string, date = new Date()): string {
  const iso = date.toISOString().slice(0, 10);
  return `${baseName}_${iso}.csv`;
}

/** Trigger a CSV file download in the browser. */
export function downloadCsv(csv: string, filename: string): void {
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
