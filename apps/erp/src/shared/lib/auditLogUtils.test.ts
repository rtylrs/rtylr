import { describe, it, expect } from 'vitest';
import {
  auditDescription,
  filterByEntity,
  filterByAction,
  filterByUser,
  filterByDateRange,
  sortByTimestamp,
  countByAction,
  latestPerEntity,
} from './auditLogUtils';
import type { AuditEntry } from './auditLogUtils';

const makeEntry = (overrides: Partial<AuditEntry> = {}): AuditEntry => ({
  id: 1,
  action: 'update',
  entityType: 'product',
  entityId: 10,
  entityName: 'Widget',
  userId: 1,
  userName: 'Alice',
  timestamp: '2026-05-15T10:00:00Z',
  ...overrides,
});

describe('auditDescription', () => {
  it('returns a readable description', () => {
    const desc = auditDescription(
      makeEntry({ action: 'create', entityType: 'product', entityName: 'Gadget', userName: 'Bob' })
    );
    expect(desc).toBe('Bob created product "Gadget"');
  });

  it('replaces underscores with spaces in entity type', () => {
    const desc = auditDescription(
      makeEntry({ entityType: 'purchase_order', entityName: 'PO-001' })
    );
    expect(desc).toContain('purchase order');
  });
});

describe('filterByEntity', () => {
  it('returns only entries of specified type', () => {
    const entries = [
      makeEntry({ entityType: 'product' }),
      makeEntry({ entityType: 'supplier' }),
      makeEntry({ entityType: 'product' }),
    ];
    expect(filterByEntity(entries, 'product')).toHaveLength(2);
    expect(filterByEntity(entries, 'supplier')).toHaveLength(1);
  });
});

describe('filterByAction', () => {
  it('filters by action', () => {
    const entries = [
      makeEntry({ action: 'create' }),
      makeEntry({ action: 'update' }),
      makeEntry({ action: 'create' }),
    ];
    expect(filterByAction(entries, 'create')).toHaveLength(2);
    expect(filterByAction(entries, 'delete')).toHaveLength(0);
  });
});

describe('filterByUser', () => {
  it('returns entries for a specific user', () => {
    const entries = [makeEntry({ userId: 1 }), makeEntry({ userId: 2 }), makeEntry({ userId: 1 })];
    expect(filterByUser(entries, 1)).toHaveLength(2);
  });
});

describe('filterByDateRange', () => {
  it('filters entries within range', () => {
    const entries = [
      makeEntry({ timestamp: '2026-05-01T00:00:00Z' }),
      makeEntry({ timestamp: '2026-05-15T00:00:00Z' }),
      makeEntry({ timestamp: '2026-06-01T00:00:00Z' }),
    ];
    const filtered = filterByDateRange(entries, '2026-05-10T00:00:00Z', '2026-05-20T00:00:00Z');
    expect(filtered).toHaveLength(1);
  });
});

describe('sortByTimestamp', () => {
  it('sorts entries most recent first', () => {
    const entries = [
      makeEntry({ timestamp: '2026-05-01T00:00:00Z' }),
      makeEntry({ timestamp: '2026-05-20T00:00:00Z' }),
      makeEntry({ timestamp: '2026-05-10T00:00:00Z' }),
    ];
    const sorted = sortByTimestamp(entries);
    expect(sorted[0].timestamp).toBe('2026-05-20T00:00:00Z');
    expect(sorted[2].timestamp).toBe('2026-05-01T00:00:00Z');
  });
});

describe('countByAction', () => {
  it('counts entries per action', () => {
    const entries = [
      makeEntry({ action: 'create' }),
      makeEntry({ action: 'update' }),
      makeEntry({ action: 'create' }),
      makeEntry({ action: 'delete' }),
    ];
    const counts = countByAction(entries);
    expect(counts.create).toBe(2);
    expect(counts.update).toBe(1);
    expect(counts.delete).toBe(1);
    expect(counts.view).toBe(0);
  });
});

describe('latestPerEntity', () => {
  it('returns most recent entry for each entity', () => {
    const entries = [
      makeEntry({ entityId: 1, timestamp: '2026-05-01T00:00:00Z' }),
      makeEntry({ entityId: 1, timestamp: '2026-05-20T00:00:00Z' }),
      makeEntry({ entityId: 2, timestamp: '2026-05-10T00:00:00Z' }),
    ];
    const latest = latestPerEntity(entries);
    expect(latest).toHaveLength(2);
    const entity1 = latest.find((e) => e.entityId === 1);
    expect(entity1!.timestamp).toBe('2026-05-20T00:00:00Z');
  });
});
