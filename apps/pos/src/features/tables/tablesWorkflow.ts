import type { Table } from '@voxire/voxie/pos/table/v1/types';
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

function textOrUndefined(value: unknown): string | undefined {
  const trimmed = String(value ?? '').trim();
  return trimmed || undefined;
}

function tableStatus(value: unknown): TableStatus {
  if (value === 'occupied') return 'seated';
  if (value === 'reserved') return 'ordered';
  if (value === 'cleaning') return 'dirty';
  return statuses.includes(value as TableStatus) ? (value as TableStatus) : 'available';
}

export function mapRemoteTables(rawTables: Table[]): TableRow[] {
  return rawTables.flatMap((table) => {
    const id = positiveInteger(table.id);
    if (!id) return [];

    return [
      {
        id,
        number: textOrUndefined(table.label) ?? String(id),
        seats: positiveInteger(table.seats) ?? 4,
        status: tableStatus(table.status),
      },
    ];
  });
}
