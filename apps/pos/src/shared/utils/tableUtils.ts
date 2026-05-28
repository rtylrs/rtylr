// Pure helpers for table management logic.

import type { TableRow } from '../../features/tables/tablesWorkflow';

/** Return the count of tables by status. */
export function tableStatusCounts(tables: TableRow[]): Record<TableRow['status'], number> {
  const counts: Record<TableRow['status'], number> = {
    available: 0,
    seated: 0,
    ordered: 0,
    paying: 0,
    dirty: 0,
  };
  for (const table of tables) {
    counts[table.status] = (counts[table.status] ?? 0) + 1;
  }
  return counts;
}

/** Filter tables by status. */
export function tablesByStatus(
  tables: TableRow[],
  status: TableRow['status'],
): TableRow[] {
  return tables.filter((t) => t.status === status);
}

/** Return true if a table is available for seating. */
export function isTableAvailable(table: TableRow): boolean {
  return table.status === 'available';
}

/** Return true if a table needs to be cleaned. */
export function isTableDirty(table: TableRow): boolean {
  return table.status === 'dirty';
}

/** Sort tables by their table number (natural sort). */
export function sortTablesByNumber(tables: TableRow[]): TableRow[] {
  return [...tables].sort((a, b) =>
    a.number.localeCompare(b.number, undefined, { numeric: true }),
  );
}

/** Find a table by its number string. */
export function findTableByNumber(
  tables: TableRow[],
  number: string,
): TableRow | undefined {
  return tables.find((t) => t.number === number);
}
