import { describe, it, expect } from 'vitest';

import {
  unreadCount,
  filterByType,
  sortByTimestamp,
  partitionByRead,
  notificationBadgeText,
  notificationTypeLabel,
} from './notificationUtils';
import type { Notification } from '../stores/flowStore';

function makeNotification(overrides: Partial<Notification>): Notification {
  return {
    id: 'n1',
    type: 'assigned',
    read: false,
    title: 'Assigned',
    description: 'Task assigned to you',
    project: 'Demo',
    projectColor: '#f00',
    actor: 'Alice',
    actorInitials: 'AL',
    timestamp: '2026-05-01T12:00:00Z',
    taskTitle: 'Fix bug',
    taskId: 't1',
    ...overrides,
  } as Notification;
}

describe('unreadCount', () => {
  it('returns 0 for empty list', () => {
    expect(unreadCount([])).toBe(0);
  });

  it('counts only unread notifications', () => {
    const notifications = [
      makeNotification({ read: false }),
      makeNotification({ id: 'n2', read: true }),
      makeNotification({ id: 'n3', read: false }),
    ];
    expect(unreadCount(notifications)).toBe(2);
  });

  it('returns 0 when all are read', () => {
    expect(unreadCount([makeNotification({ read: true })])).toBe(0);
  });
});

describe('filterByType', () => {
  it('returns only notifications of the given type', () => {
    const notifications = [
      makeNotification({ id: 'a', type: 'assigned' }),
      makeNotification({ id: 'b', type: 'comment' }),
      makeNotification({ id: 'c', type: 'assigned' }),
    ];
    expect(filterByType(notifications, 'assigned').map((n) => n.id)).toEqual(['a', 'c']);
  });

  it('returns empty array when no matches', () => {
    expect(filterByType([makeNotification({ type: 'assigned' })], 'comment')).toHaveLength(0);
  });
});

describe('sortByTimestamp', () => {
  it('sorts newest first', () => {
    const notifications = [
      makeNotification({ id: 'a', timestamp: '2026-01-01T00:00:00Z' }),
      makeNotification({ id: 'b', timestamp: '2026-06-01T00:00:00Z' }),
      makeNotification({ id: 'c', timestamp: '2026-03-01T00:00:00Z' }),
    ];
    const sorted = sortByTimestamp(notifications);
    expect(sorted.map((n) => n.id)).toEqual(['b', 'c', 'a']);
  });

  it('does not mutate the original array', () => {
    const notifications = [
      makeNotification({ timestamp: '2026-06-01T00:00:00Z' }),
      makeNotification({ timestamp: '2026-01-01T00:00:00Z' }),
    ];
    const original = [...notifications];
    sortByTimestamp(notifications);
    expect(notifications).toEqual(original);
  });
});

describe('partitionByRead', () => {
  it('separates read and unread notifications', () => {
    const notifications = [
      makeNotification({ id: 'a', read: false }),
      makeNotification({ id: 'b', read: true }),
      makeNotification({ id: 'c', read: false }),
    ];
    const { read, unread } = partitionByRead(notifications);
    expect(read.map((n) => n.id)).toEqual(['b']);
    expect(unread.map((n) => n.id)).toEqual(['a', 'c']);
  });
});

describe('notificationBadgeText', () => {
  it('returns empty string for 0', () => {
    expect(notificationBadgeText(0)).toBe('');
  });

  it('returns the count as string for 1-9', () => {
    expect(notificationBadgeText(1)).toBe('1');
    expect(notificationBadgeText(9)).toBe('9');
  });

  it('returns "9+" for counts above 9', () => {
    expect(notificationBadgeText(10)).toBe('9+');
    expect(notificationBadgeText(99)).toBe('9+');
  });
});

describe('notificationTypeLabel', () => {
  it('returns correct label for assigned', () => {
    expect(notificationTypeLabel('assigned')).toBe('Assigned to you');
  });

  it('returns correct label for comment', () => {
    expect(notificationTypeLabel('comment')).toBe('New comment');
  });

  it('returns correct label for due_soon', () => {
    expect(notificationTypeLabel('due_soon')).toBe('Due soon');
  });
});
