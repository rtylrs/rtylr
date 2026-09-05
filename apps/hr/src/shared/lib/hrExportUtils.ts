import { todayIso } from './dateUtils';

export type ExportFormat = 'csv' | 'xlsx' | 'json' | 'pdf';
export type ExportScope = 'all' | 'selected' | 'filtered';

export function exportFormatLabel(format: ExportFormat): string {
  const labels: Record<ExportFormat, string> = {
    csv: 'CSV',
    xlsx: 'Excel',
    json: 'JSON',
    pdf: 'PDF',
  };
  return labels[format] ?? format.toUpperCase();
}

export function exportFormatMimeType(format: ExportFormat): string {
  const mimes: Record<ExportFormat, string> = {
    csv: 'text/csv',
    xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    json: 'application/json',
    pdf: 'application/pdf',
  };
  return mimes[format] ?? 'application/octet-stream';
}

export function exportScopeLabel(scope: ExportScope): string {
  const labels: Record<ExportScope, string> = {
    all: 'All Records',
    selected: 'Selected Records',
    filtered: 'Filtered Records',
  };
  return labels[scope] ?? scope;
}

export function generateExportFilename(
  entityName: string,
  format: ExportFormat,
  dateIso?: string
): string {
  const date = dateIso ?? todayIso();
  const slug = entityName.toLowerCase().replace(/\s+/g, '_');
  return `${slug}_export_${date}.${format}`;
}

export function objectsToCSV<T extends Record<string, unknown>>(
  objects: T[],
  headers: Array<keyof T>
): string {
  if (objects.length === 0) return '';
  const headerRow = headers.map(String).join(',');
  const rows = objects.map((obj) =>
    headers
      .map((h) => {
        const val = String(obj[h] ?? '');
        return val.includes(',') ? `"${val}"` : val;
      })
      .join(',')
  );
  return [headerRow, ...rows].join('\n');
}

export function estimateExportSizeKB(recordCount: number, fieldCount: number): number {
  const avgBytesPerField = 20;
  return Math.ceil((recordCount * fieldCount * avgBytesPerField) / 1024);
}
