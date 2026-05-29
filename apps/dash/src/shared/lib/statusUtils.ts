/**
 * Status display utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type RecordStatus = 'active' | 'inactive' | 'pending' | 'suspended' | 'archived';

export type StatusVariant = 'success' | 'warning' | 'error' | 'neutral' | 'info';

const STATUS_LABELS: Record<RecordStatus, string> = {
  active: 'Active',
  inactive: 'Inactive',
  pending: 'Pending',
  suspended: 'Suspended',
  archived: 'Archived',
};

const STATUS_VARIANTS: Record<RecordStatus, StatusVariant> = {
  active: 'success',
  inactive: 'neutral',
  pending: 'warning',
  suspended: 'error',
  archived: 'neutral',
};

/** Return the human-readable label for a record status. */
export function statusLabel(status: RecordStatus): string {
  return STATUS_LABELS[status] ?? status;
}

/** Return the badge variant for a record status. */
export function statusVariant(status: RecordStatus): StatusVariant {
  return STATUS_VARIANTS[status] ?? 'neutral';
}

/** Return true when the status is considered active. */
export function isActiveStatus(status: RecordStatus): boolean {
  return status === 'active';
}

/** Return true when the status is terminal (no further transitions expected). */
export function isTerminalStatus(status: RecordStatus): boolean {
  return status === 'archived' || status === 'suspended';
}

/** Return the list of statuses a record can transition to from the current one. */
export function allowedTransitions(status: RecordStatus): RecordStatus[] {
  switch (status) {
    case 'active':
      return ['inactive', 'suspended', 'archived'];
    case 'inactive':
      return ['active', 'archived'];
    case 'pending':
      return ['active', 'suspended', 'archived'];
    case 'suspended':
      return ['active', 'archived'];
    case 'archived':
      return [];
    default:
      return [];
  }
}

/** Return true when transitioning from `from` to `to` is allowed. */
export function canTransition(from: RecordStatus, to: RecordStatus): boolean {
  return allowedTransitions(from).includes(to);
}

/** Return all non-terminal statuses. */
export function activeStatuses(): RecordStatus[] {
  return (['active', 'inactive', 'pending'] as RecordStatus[]).filter((s) => !isTerminalStatus(s));
}
