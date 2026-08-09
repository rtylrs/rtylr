// CRM contact/record status helpers.

export type CrmStatus = 'active' | 'inactive' | 'pending' | 'deleted';

const STATUS_LABELS: Record<CrmStatus, string> = {
  active: 'Active',
  inactive: 'Inactive',
  pending: 'Pending',
  deleted: 'Deleted',
};

const STATUS_COLORS: Record<CrmStatus, string> = {
  active: 'green',
  inactive: 'gray',
  pending: 'yellow',
  deleted: 'red',
};

const STATUS_BADGE_CLASSES: Record<CrmStatus, string> = {
  active: 'bg-green-100 text-green-800',
  inactive: 'bg-gray-100 text-gray-800',
  pending: 'bg-blue-100 text-blue-800',
  deleted: 'bg-red-100 text-red-800',
};

const ALLOWED_TRANSITIONS: Record<CrmStatus, CrmStatus[]> = {
  active: ['inactive', 'deleted'],
  inactive: ['active', 'deleted'],
  pending: ['active', 'deleted'],
  deleted: [],
};

/**
 * statusLabel returns a human-readable label for a status.
 */
export function statusLabel(status: CrmStatus | string): string {
  return (STATUS_LABELS as Record<string, string>)[status] ?? status;
}

/**
 * statusColor returns a color token for a status.
 */
export function statusColor(status: CrmStatus | string): string {
  return (STATUS_COLORS as Record<string, string>)[status] ?? 'gray';
}

/**
 * statusBadgeClass returns Tailwind badge classes for a status.
 */
export function statusBadgeClass(status: CrmStatus | string): string {
  return (STATUS_BADGE_CLASSES as Record<string, string>)[status] ?? 'bg-gray-100 text-gray-800';
}

/**
 * isUsable returns true when the status allows normal usage.
 */
export function isUsable(status: CrmStatus | string): boolean {
  return status === 'active' || status === 'pending';
}

/**
 * allowedTransitions returns the valid next statuses from the given status.
 */
export function allowedTransitions(status: CrmStatus | string): CrmStatus[] {
  return (ALLOWED_TRANSITIONS as Record<string, CrmStatus[]>)[status] ?? [];
}

/**
 * canTransition returns true when transitioning from one status to another is allowed.
 */
export function canTransition(from: CrmStatus | string, to: CrmStatus | string): boolean {
  return allowedTransitions(from).includes(to as CrmStatus);
}
