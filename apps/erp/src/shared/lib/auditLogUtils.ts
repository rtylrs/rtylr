/**
 * Audit log helpers for tracking ERP data changes.
 * All functions are pure with no side-effects.
 */

export type AuditAction = 'create' | 'update' | 'delete' | 'view' | 'export';
export type AuditEntityType =
  | 'product'
  | 'category'
  | 'brand'
  | 'supplier'
  | 'purchase_order'
  | 'stock_take'
  | 'stock_transfer'
  | 'warehouse'
  | 'price_list';

export interface AuditEntry {
  id: number;
  action: AuditAction;
  entityType: AuditEntityType;
  entityId: number;
  entityName: string;
  userId: number;
  userName: string;
  timestamp: string;
  changes?: Record<string, { from: unknown; to: unknown }>;
}

/** Return a human-readable description of an audit entry. */
export function auditDescription(entry: AuditEntry): string {
  const actionLabel: Record<AuditAction, string> = {
    create: 'created',
    update: 'updated',
    delete: 'deleted',
    view: 'viewed',
    export: 'exported',
  };
  const entityLabel = entry.entityType.replace(/_/g, ' ');
  return `${entry.userName} ${actionLabel[entry.action]} ${entityLabel} "${entry.entityName}"`;
}

/** Filter audit entries by entity type. */
export function filterByEntity(entries: AuditEntry[], entityType: AuditEntityType): AuditEntry[] {
  return entries.filter((e) => e.entityType === entityType);
}

/** Filter audit entries by action. */
export function filterByAction(entries: AuditEntry[], action: AuditAction): AuditEntry[] {
  return entries.filter((e) => e.action === action);
}

/** Filter audit entries by user id. */
export function filterByUser(entries: AuditEntry[], userId: number): AuditEntry[] {
  return entries.filter((e) => e.userId === userId);
}

/** Filter audit entries within a date range (ISO timestamp strings). */
export function filterByDateRange(entries: AuditEntry[], from: string, to: string): AuditEntry[] {
  return entries.filter((e) => e.timestamp >= from && e.timestamp <= to);
}

/** Sort audit entries by timestamp descending. */
export function sortByTimestamp(entries: AuditEntry[]): AuditEntry[] {
  return [...entries].sort((a, b) => b.timestamp.localeCompare(a.timestamp));
}

/** Count entries per action type. */
export function countByAction(entries: AuditEntry[]): Record<AuditAction, number> {
  const counts: Record<AuditAction, number> = {
    create: 0,
    update: 0,
    delete: 0,
    view: 0,
    export: 0,
  };
  for (const entry of entries) {
    counts[entry.action]++;
  }
  return counts;
}

/** Return the most recent entry for each entity (by entityType + entityId). */
export function latestPerEntity(entries: AuditEntry[]): AuditEntry[] {
  const map = new Map<string, AuditEntry>();
  for (const entry of sortByTimestamp(entries)) {
    const key = `${entry.entityType}:${entry.entityId}`;
    if (!map.has(key)) {
      map.set(key, entry);
    }
  }
  return Array.from(map.values());
}
