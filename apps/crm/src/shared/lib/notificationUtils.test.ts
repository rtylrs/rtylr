import { describe, it, expect } from 'vitest';
import {
  isUnread,
  unreadCount,
  markAllRead,
  markRead,
  filterByType,
  notificationIcon,
  sortByDate,
} from './notificationUtils';
import type { Notification } from './notificationUtils';

const notifications: Notification[] = [
  { id: 'n1', type: 'info', title: 'Update', read: false, createdAt: '2024-01-03T10:00:00Z' },
  { id: 'n2', type: 'success', title: 'Done', read: true, createdAt: '2024-01-01T10:00:00Z' },
  { id: 'n3', type: 'error', title: 'Fail', read: false, createdAt: '2024-01-02T10:00:00Z' },
];

describe('isUnread', () => {
  it('returns true when not read', () => expect(isUnread(notifications[0])).toBe(true));
  it('returns false when read', () => expect(isUnread(notifications[1])).toBe(false));
});

describe('unreadCount', () => {
  it('counts unread notifications', () => expect(unreadCount(notifications)).toBe(2));
});

describe('markAllRead', () => {
  it('marks all as read', () => {
    const result = markAllRead(notifications);
    expect(result.every((n) => n.read)).toBe(true);
  });

  it('does not mutate original', () => {
    markAllRead(notifications);
    expect(notifications[0].read).toBe(false);
  });
});

describe('markRead', () => {
  it('marks a single notification as read', () => {
    const result = markRead(notifications, 'n1');
    expect(result.find((n) => n.id === 'n1')?.read).toBe(true);
    expect(result.find((n) => n.id === 'n3')?.read).toBe(false);
  });
});

describe('filterByType', () => {
  it('filters by notification type', () => {
    expect(filterByType(notifications, 'info')).toHaveLength(1);
    expect(filterByType(notifications, 'warning')).toHaveLength(0);
  });
});

describe('notificationIcon', () => {
  it('returns correct icon names', () => {
    expect(notificationIcon('info')).toBe('info-circle');
    expect(notificationIcon('success')).toBe('check-circle');
    expect(notificationIcon('error')).toBe('x-circle');
  });
});

describe('sortByDate', () => {
  it('sorts newest first', () => {
    const sorted = sortByDate(notifications);
    expect(sorted[0].id).toBe('n1');
    expect(sorted[2].id).toBe('n2');
  });

  it('does not mutate original', () => {
    const copy = [...notifications];
    sortByDate(notifications);
    expect(notifications).toEqual(copy);
  });
});
