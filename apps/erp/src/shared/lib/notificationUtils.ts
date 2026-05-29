/**
 * ERP notification helpers for in-app alerts and system events.
 * All functions are pure with no side-effects.
 */

export type NotificationType =
  | 'low_stock'
  | 'order_received'
  | 'order_approved'
  | 'transfer_complete'
  | 'expiry_warning'
  | 'system';

export type NotificationChannel = 'in_app' | 'email' | 'sms';

export interface ErpNotification {
  id: number;
  type: NotificationType;
  title: string;
  message: string;
  channel: NotificationChannel;
  read: boolean;
  createdAt: string;
}

/** Return an icon label for a notification type. */
export function notificationIcon(type: NotificationType): string {
  const icons: Record<NotificationType, string> = {
    low_stock: 'alert-triangle',
    order_received: 'package',
    order_approved: 'check-circle',
    transfer_complete: 'truck',
    expiry_warning: 'clock',
    system: 'bell',
  };
  return icons[type];
}

/** Return a color class for a notification type. */
export function notificationColor(type: NotificationType): string {
  const colors: Record<NotificationType, string> = {
    low_stock: 'text-red-600',
    order_received: 'text-blue-600',
    order_approved: 'text-green-600',
    transfer_complete: 'text-purple-600',
    expiry_warning: 'text-amber-600',
    system: 'text-gray-600',
  };
  return colors[type];
}

/** Count unread notifications. */
export function unreadCount(notifications: ErpNotification[]): number {
  return notifications.filter((n) => !n.read).length;
}

/** Return only unread notifications. */
export function unreadNotifications(notifications: ErpNotification[]): ErpNotification[] {
  return notifications.filter((n) => !n.read);
}

/** Mark all notifications as read (returns new array). */
export function markAllRead(notifications: ErpNotification[]): ErpNotification[] {
  return notifications.map((n) => ({ ...n, read: true }));
}

/** Filter notifications by type. */
export function filterByType(
  notifications: ErpNotification[],
  type: NotificationType
): ErpNotification[] {
  return notifications.filter((n) => n.type === type);
}

/** Sort notifications most recent first. */
export function sortByRecent(notifications: ErpNotification[]): ErpNotification[] {
  return [...notifications].sort((a, b) => b.createdAt.localeCompare(a.createdAt));
}

/** Group notifications by type. */
export function groupByType(
  notifications: ErpNotification[]
): Partial<Record<NotificationType, ErpNotification[]>> {
  const groups: Partial<Record<NotificationType, ErpNotification[]>> = {};
  for (const n of notifications) {
    if (!groups[n.type]) {
      groups[n.type] = [];
    }
    groups[n.type]!.push(n);
  }
  return groups;
}
