import { describe, it, expect } from 'vitest';
import {
  notificationIcon,
  notificationColor,
  unreadCount,
  unreadNotifications,
  markAllRead,
  filterByType,
  sortByRecent,
  groupByType,
} from './notificationUtils';
import type { ErpNotification } from './notificationUtils';

const makeNotification = (overrides: Partial<ErpNotification> = {}): ErpNotification => ({
  id: 1,
  type: 'system',
  title: 'Test',
  message: 'Test message',
  channel: 'in_app',
  read: false,
  createdAt: '2026-05-15T10:00:00Z',
  ...overrides,
});

describe('notificationIcon', () => {
  it('returns correct icon for each type', () => {
    expect(notificationIcon('low_stock')).toBe('alert-triangle');
    expect(notificationIcon('order_approved')).toBe('check-circle');
    expect(notificationIcon('system')).toBe('bell');
  });
});

describe('notificationColor', () => {
  it('returns a color class string', () => {
    expect(notificationColor('low_stock')).toContain('red');
    expect(notificationColor('order_approved')).toContain('green');
    expect(notificationColor('expiry_warning')).toContain('amber');
  });
});

describe('unreadCount', () => {
  it('counts unread notifications', () => {
    const notifications = [
      makeNotification({ read: false }),
      makeNotification({ read: true }),
      makeNotification({ read: false }),
    ];
    expect(unreadCount(notifications)).toBe(2);
  });

  it('returns 0 when all read', () => {
    expect(unreadCount([makeNotification({ read: true })])).toBe(0);
  });
});

describe('unreadNotifications', () => {
  it('filters to only unread', () => {
    const notifications = [
      makeNotification({ id: 1, read: false }),
      makeNotification({ id: 2, read: true }),
    ];
    const result = unreadNotifications(notifications);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(1);
  });
});

describe('markAllRead', () => {
  it('marks all as read', () => {
    const notifications = [makeNotification({ read: false }), makeNotification({ read: false })];
    const result = markAllRead(notifications);
    expect(result.every((n) => n.read)).toBe(true);
  });

  it('does not mutate the original array', () => {
    const notifications = [makeNotification({ read: false })];
    markAllRead(notifications);
    expect(notifications[0].read).toBe(false);
  });
});

describe('filterByType', () => {
  it('returns only notifications of given type', () => {
    const notifications = [
      makeNotification({ type: 'low_stock' }),
      makeNotification({ type: 'system' }),
      makeNotification({ type: 'low_stock' }),
    ];
    expect(filterByType(notifications, 'low_stock')).toHaveLength(2);
    expect(filterByType(notifications, 'order_received')).toHaveLength(0);
  });
});

describe('sortByRecent', () => {
  it('sorts newest first', () => {
    const notifications = [
      makeNotification({ createdAt: '2026-05-01T00:00:00Z' }),
      makeNotification({ createdAt: '2026-05-20T00:00:00Z' }),
      makeNotification({ createdAt: '2026-05-10T00:00:00Z' }),
    ];
    const sorted = sortByRecent(notifications);
    expect(sorted[0].createdAt).toBe('2026-05-20T00:00:00Z');
    expect(sorted[2].createdAt).toBe('2026-05-01T00:00:00Z');
  });
});

describe('groupByType', () => {
  it('groups notifications by type', () => {
    const notifications = [
      makeNotification({ type: 'system' }),
      makeNotification({ type: 'low_stock' }),
      makeNotification({ type: 'system' }),
    ];
    const groups = groupByType(notifications);
    expect(groups.system).toHaveLength(2);
    expect(groups.low_stock).toHaveLength(1);
  });
});
