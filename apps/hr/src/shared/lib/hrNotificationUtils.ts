export type NotificationType =
  | 'leave_request'
  | 'leave_approved'
  | 'leave_rejected'
  | 'payroll_processed'
  | 'review_due'
  | 'birthday'
  | 'anniversary'
  | 'system';

export type NotificationPriority = 'low' | 'medium' | 'high' | 'urgent';

export function notificationTypeLabel(type: NotificationType): string {
  const labels: Record<NotificationType, string> = {
    leave_request: 'Leave Request',
    leave_approved: 'Leave Approved',
    leave_rejected: 'Leave Rejected',
    payroll_processed: 'Payroll Processed',
    review_due: 'Review Due',
    birthday: 'Birthday',
    anniversary: 'Work Anniversary',
    system: 'System',
  };
  return labels[type] ?? type;
}

export function notificationPriorityLabel(priority: NotificationPriority): string {
  const labels: Record<NotificationPriority, string> = {
    low: 'Low',
    medium: 'Medium',
    high: 'High',
    urgent: 'Urgent',
  };
  return labels[priority] ?? priority;
}

export function notificationPriorityColor(priority: NotificationPriority): string {
  const colors: Record<NotificationPriority, string> = {
    low: 'gray',
    medium: 'blue',
    high: 'orange',
    urgent: 'red',
  };
  return colors[priority] ?? 'gray';
}

export function isActionRequired(type: NotificationType): boolean {
  return type === 'leave_request' || type === 'review_due';
}

export function countUnread(notifications: Array<{ read: boolean }>): number {
  return notifications.filter((n) => !n.read).length;
}

export function filterByPriority<T extends { priority: NotificationPriority }>(
  notifications: T[],
  priority: NotificationPriority
): T[] {
  return notifications.filter((n) => n.priority === priority);
}

const PRIORITY_ORDER: Record<NotificationPriority, number> = {
  urgent: 4,
  high: 3,
  medium: 2,
  low: 1,
};

export function sortByPriorityDesc<T extends { priority: NotificationPriority }>(
  notifications: T[]
): T[] {
  return [...notifications].sort((a, b) => PRIORITY_ORDER[b.priority] - PRIORITY_ORDER[a.priority]);
}
