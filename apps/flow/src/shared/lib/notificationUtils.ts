import type { Notification } from '../stores/flowStore';

/** Returns the count of unread notifications. */
export function unreadCount(notifications: Notification[]): number {
  return notifications.filter((n) => !n.read).length;
}

/** Filter notifications by type. */
export function filterByType(
  notifications: Notification[],
  type: Notification['type']
): Notification[] {
  return notifications.filter((n) => n.type === type);
}

/** Sort notifications by timestamp (newest first). */
export function sortByTimestamp(notifications: Notification[]): Notification[] {
  return [...notifications].sort(
    (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
  );
}

/** Group notifications into read and unread buckets. */
export function partitionByRead(notifications: Notification[]): {
  read: Notification[];
  unread: Notification[];
} {
  const read: Notification[] = [];
  const unread: Notification[] = [];
  for (const n of notifications) {
    if (n.read) read.push(n);
    else unread.push(n);
  }
  return { read, unread };
}

/** Returns a display badge string: "9+" if count > 9, else the count. */
export function notificationBadgeText(count: number): string {
  if (count <= 0) return '';
  return count > 9 ? '9+' : String(count);
}

/** Human-friendly label for a notification type. */
export function notificationTypeLabel(type: Notification['type']): string {
  const labels: Record<Notification['type'], string> = {
    assigned: 'Assigned to you',
    mention: 'Mentioned you',
    comment: 'New comment',
    status_change: 'Status changed',
    due_soon: 'Due soon',
  };
  return labels[type] ?? type;
}
