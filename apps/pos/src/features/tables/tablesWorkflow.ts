import type { TableStatus } from '../../shared/components/TableCard';

export interface TableRow {
  id: number;
  number: string;
  seats: number;
  status: TableStatus;
  guests?: number;
  total?: number;
  server?: string;
}

const statuses: TableStatus[] = ['available', 'seated', 'ordered', 'paying', 'dirty'];

function positiveInteger(value: unknown): number | null {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) return null;
  return Math.floor(numeric);
}

function optionalNonNegativeInteger(value: unknown): number | undefined {
  if (value === undefined || value === null || value === '') return undefined;
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric < 0) return undefined;
  return Math.floor(numeric);
}

function nonNegativeNumber(value: unknown): number | undefined {
  if (value === undefined || value === null || value === '') return undefined;
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) return undefined;
  return Math.max(0, numeric);
}

function textOrUndefined(value: unknown): string | undefined {
  const trimmed = String(value ?? '').trim();
  return trimmed || undefined;
}

function record(value: unknown): Record<string, unknown> | null {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

function tableStatus(value: unknown): TableStatus {
  if (value === 'occupied') return 'seated';
  if (value === 'reserved') return 'ordered';
  if (value === 'cleaning') return 'dirty';
  return statuses.includes(value as TableStatus) ? (value as TableStatus) : 'available';
}

export function mapRemoteTables(rawTables: unknown[]): TableRow[] {
  return rawTables.flatMap((table) => {
    const raw = record(table);
    if (!raw) return [];

    const id = positiveInteger(raw.id);
    if (!id) return [];

    return [
      {
        id,
        number: textOrUndefined(raw.number ?? raw.label ?? raw.name) ?? String(id),
        seats: positiveInteger(raw.seats) ?? 4,
        status: tableStatus(raw.status),
        guests: optionalNonNegativeInteger(raw.guests),
        total: nonNegativeNumber(raw.total),
        server: textOrUndefined(raw.server),
      },
    ];
  });
}

function remoteTablesArray(response: unknown): unknown[] {
  if (Array.isArray(response)) return response;

  const root = record(response);
  if (!root) return [];
  if (Array.isArray(root.tables)) return root.tables;

  const data = record(root.data);
  if (data && Array.isArray(data.tables)) return data.tables;
  if (data && Array.isArray(data.data)) return data.data;
  if (Array.isArray(root.data)) return root.data;

  return [];
}

export function normalizeTablesResponse(response: unknown): TableRow[] {
  return mapRemoteTables(remoteTablesArray(response));
}
