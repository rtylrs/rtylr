// Application status helpers for the recruit admin panel.

export type ApplicationStatus =
  | 'new'
  | 'reviewing'
  | 'interview'
  | 'offer'
  | 'hired'
  | 'rejected'
  | 'withdrawn';

const STATUS_ORDER: ApplicationStatus[] = [
  'new',
  'reviewing',
  'interview',
  'offer',
  'hired',
  'rejected',
  'withdrawn',
];

const STATUS_COLORS: Record<ApplicationStatus, string> = {
  new: 'blue',
  reviewing: 'yellow',
  interview: 'purple',
  offer: 'orange',
  hired: 'green',
  rejected: 'red',
  withdrawn: 'gray',
};

const STATUS_LABELS: Record<ApplicationStatus, string> = {
  new: 'New',
  reviewing: 'Reviewing',
  interview: 'Interview',
  offer: 'Offer Made',
  hired: 'Hired',
  rejected: 'Rejected',
  withdrawn: 'Withdrawn',
};

/**
 * allStatuses returns all application statuses in pipeline order.
 */
export function allStatuses(): ApplicationStatus[] {
  return [...STATUS_ORDER];
}

/**
 * statusColor returns a color token for the given status.
 */
export function statusColor(status: ApplicationStatus | string): string {
  return (STATUS_COLORS as Record<string, string>)[status] ?? 'gray';
}

/**
 * statusLabel returns a human-readable label for the given status.
 */
export function statusLabel(status: ApplicationStatus | string): string {
  return (STATUS_LABELS as Record<string, string>)[status] ?? status;
}

/**
 * statusIndex returns the pipeline position index (0-based).
 * Returns -1 for unknown statuses.
 */
export function statusIndex(status: ApplicationStatus | string): number {
  return STATUS_ORDER.indexOf(status as ApplicationStatus);
}

/**
 * isTerminalStatus returns true for hired, rejected, and withdrawn.
 */
export function isTerminalStatus(status: ApplicationStatus | string): boolean {
  return status === 'hired' || status === 'rejected' || status === 'withdrawn';
}

/**
 * isActiveStatus returns true when the application is still progressing.
 */
export function isActiveStatus(status: ApplicationStatus | string): boolean {
  return !isTerminalStatus(status);
}
