/**
 * Notification utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type NotificationSeverity = 'info' | 'success' | 'warning' | 'error';
export type NotificationChannel = 'email' | 'push' | 'sms' | 'in_app';

export interface Notification {
  id: string;
  title: string;
  body: string;
  severity: NotificationSeverity;
  read: boolean;
  createdAt: string;
}

/** Return a display label for a notification severity. */
export function severityLabel(severity: NotificationSeverity): string {
  const labels: Record<NotificationSeverity, string> = {
    info: 'Info',
    success: 'Success',
    warning: 'Warning',
    error: 'Error',
  };
  return labels[severity] ?? severity;
}

/** Return true if the severity requires immediate attention. */
export function isUrgent(severity: NotificationSeverity): boolean {
  return severity === 'error' || severity === 'warning';
}

/** Filter notifications to only unread ones. */
export function unreadNotifications(notifications: Notification[]): Notification[] {
  return notifications.filter((n) => !n.read);
}

/** Count unread notifications. */
export function unreadCount(notifications: Notification[]): number {
  return notifications.filter((n) => !n.read).length;
}

/** Sort notifications by createdAt descending (newest first). */
export function sortByDate(notifications: Notification[]): Notification[] {
  return [...notifications].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

/** Return the notification badge label (e.g. "9+" when count > maxDisplay). */
export function badgeLabel(count: number, maxDisplay = 9): string {
  if (count <= 0) return '';
  return count > maxDisplay ? `${maxDisplay}+` : String(count);
}

/** Return true if a channel is enabled in a given set of enabled channels. */
export function isChannelEnabled(
  channel: NotificationChannel,
  enabled: NotificationChannel[]
): boolean {
  return enabled.includes(channel);
}
