import type { Notification } from '../stores/flowStore';

/** Returns unread notifications. */
export function unreadNotifications(notifications: Notification[]): Notification[] {
  return notifications.filter((n) => !n.read);
}

/** Returns read notifications. */
export function readNotifications(notifications: Notification[]): Notification[] {
  return notifications.filter((n) => n.read);
}

/** Returns the count of unread notifications. */
export function unreadCount(notifications: Notification[]): number {
  return unreadNotifications(notifications).length;
}

/** Returns notifications of a specific type. */
export function notificationsByType(
  notifications: Notification[],
  type: Notification['type']
): Notification[] {
  return notifications.filter((n) => n.type === type);
}

/** Returns notifications sorted by timestamp descending (newest first). */
export function sortNotificationsDesc(notifications: Notification[]): Notification[] {
  return [...notifications].sort((a, b) => b.timestamp.localeCompare(a.timestamp));
}

/** Returns notifications for a specific task. */
export function notificationsForTask(
  notifications: Notification[],
  taskId: string
): Notification[] {
  return notifications.filter((n) => n.taskId === taskId);
}

/** Returns a display label for a notification type. */
export function notificationTypeLabel(type: Notification['type']): string {
  const labels: Record<Notification['type'], string> = {
    assigned: 'Assigned to you',
    mention: 'Mentioned you',
    comment: 'New comment',
    status_change: 'Status changed',
    due_soon: 'Due soon',
  };
  return labels[type];
}

/** Returns true if any notification is unread. */
export function hasUnreadNotification(notifications: Notification[]): boolean {
  return notifications.some((n) => !n.read);
}
