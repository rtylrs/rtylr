export type AuditAction =
  'created' | 'updated' | 'deleted' | 'exported' | 'imported' | 'sent' | 'paused' | 'resumed';

export interface AuditEntry {
  id: number;
  entityType: string;
  entityId: number;
  action: AuditAction;
  performedBy: string;
  performedAt: string;
  details?: string;
}

export function actionLabel(action: AuditAction): string {
  const labels: Record<AuditAction, string> = {
    created: 'Created',
    updated: 'Updated',
    deleted: 'Deleted',
    exported: 'Exported',
    imported: 'Imported',
    sent: 'Sent',
    paused: 'Paused',
    resumed: 'Resumed',
  };
  return labels[action];
}

export function filterByEntity(entries: AuditEntry[], entityType: string): AuditEntry[] {
  return entries.filter((e) => e.entityType === entityType);
}

export function filterByPerformer(entries: AuditEntry[], performedBy: string): AuditEntry[] {
  return entries.filter((e) => e.performedBy === performedBy);
}

export function filterByAction(entries: AuditEntry[], action: AuditAction): AuditEntry[] {
  return entries.filter((e) => e.action === action);
}

export function sortNewest(entries: AuditEntry[]): AuditEntry[] {
  return [...entries].sort((a, b) => b.performedAt.localeCompare(a.performedAt));
}

export function groupByEntityType(entries: AuditEntry[]): Record<string, AuditEntry[]> {
  const groups: Record<string, AuditEntry[]> = {};
  for (const entry of entries) {
    (groups[entry.entityType] ??= []).push(entry);
  }
  return groups;
}
