import { describe, it, expect } from 'vitest';
import {
  tableStatusCounts,
  tablesByStatus,
  isTableAvailable,
  isTableDirty,
  sortTablesByNumber,
  findTableByNumber,
} from './tableUtils';
import type { TableRow } from '../../features/tables/tablesWorkflow';

const makeTable = (overrides: Partial<TableRow> = {}): TableRow => ({
  id: 1,
  number: 'T1',
  seats: 4,
  status: 'available',
  ...overrides,
});

const tables: TableRow[] = [
  makeTable({ id: 1, number: 'T1', status: 'available' }),
  makeTable({ id: 2, number: 'T2', status: 'seated' }),
  makeTable({ id: 3, number: 'T3', status: 'ordered' }),
  makeTable({ id: 4, number: 'T10', status: 'available' }),
  makeTable({ id: 5, number: 'T5', status: 'dirty' }),
];

describe('tableStatusCounts', () => {
  it('counts each status correctly', () => {
    const counts = tableStatusCounts(tables);
    expect(counts.available).toBe(2);
    expect(counts.seated).toBe(1);
    expect(counts.ordered).toBe(1);
    expect(counts.dirty).toBe(1);
    expect(counts.paying).toBe(0);
  });

  it('returns all zeros for an empty list', () => {
    const counts = tableStatusCounts([]);
    expect(counts.available).toBe(0);
    expect(counts.seated).toBe(0);
  });
});

describe('tablesByStatus', () => {
  it('filters tables by status', () => {
    const available = tablesByStatus(tables, 'available');
    expect(available).toHaveLength(2);
  });

  it('returns an empty array when no tables match', () => {
    expect(tablesByStatus(tables, 'paying')).toHaveLength(0);
  });
});

describe('isTableAvailable', () => {
  it('returns true for an available table', () => {
    expect(isTableAvailable(makeTable({ status: 'available' }))).toBe(true);
  });

  it('returns false for a seated table', () => {
    expect(isTableAvailable(makeTable({ status: 'seated' }))).toBe(false);
  });
});

describe('isTableDirty', () => {
  it('returns true for a dirty table', () => {
    expect(isTableDirty(makeTable({ status: 'dirty' }))).toBe(true);
  });

  it('returns false for an available table', () => {
    expect(isTableDirty(makeTable({ status: 'available' }))).toBe(false);
  });
});

describe('sortTablesByNumber', () => {
  it('sorts tables numerically by number', () => {
    const sorted = sortTablesByNumber(tables);
    expect(sorted.map((t) => t.number)).toEqual(['T1', 'T2', 'T3', 'T5', 'T10']);
  });

  it('does not mutate the original array', () => {
    const original = [...tables];
    sortTablesByNumber(tables);
    expect(tables).toEqual(original);
  });
});

describe('findTableByNumber', () => {
  it('finds a table by its number', () => {
    const result = findTableByNumber(tables, 'T3');
    expect(result).toBeDefined();
    expect(result?.id).toBe(3);
  });

  it('returns undefined when no table has the given number', () => {
    expect(findTableByNumber(tables, 'T99')).toBeUndefined();
  });
});
