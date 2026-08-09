import { describe, it, expect } from 'vitest';

import {
  unreadNotifications,
  readNotifications,
  unreadCount,
  notificationsByType,
  sortNotificationsDesc,
  notificationsForTask,
  notificationTypeLabel,
  hasUnreadNotification,
} from './notificationFilterUtils';
import type { Notification } from '../stores/flowStore';

function makeNotification(overrides: Partial<Notification>): Notification {
  return {
    id: 'n1',
    type: 'assigned',
    read: false,
    title: 'Assigned to you',
    description: 'Task was assigned',
    project: 'Auth',
    projectColor: '#3b82f6',
    actor: 'Alice',
    actorInitials: 'A',
    timestamp: '2026-06-15T10:00:00Z',
    taskTitle: 'Fix bug',
    taskId: 't1',
    ...overrides,
  } as Notification;
}

describe('unreadNotifications', () => {
  it('returns only unread notifications', () => {
    const notifications = [
      makeNotification({ read: false }),
      makeNotification({ id: 'n2', read: true }),
    ];
    expect(unreadNotifications(notifications).length).toBe(1);
  });
});

describe('readNotifications', () => {
  it('returns only read notifications', () => {
    const notifications = [
      makeNotification({ read: false }),
      makeNotification({ id: 'n2', read: true }),
    ];
    expect(readNotifications(notifications).length).toBe(1);
  });
});

describe('unreadCount', () => {
  it('counts unread notifications', () => {
    const notifications = [
      makeNotification({ read: false }),
      makeNotification({ id: 'n2', read: false }),
      makeNotification({ id: 'n3', read: true }),
    ];
    expect(unreadCount(notifications)).toBe(2);
  });
});

describe('notificationsByType', () => {
  it('filters by type', () => {
    const notifications = [
      makeNotification({ type: 'assigned' }),
      makeNotification({ id: 'n2', type: 'mention' }),
    ];
    expect(notificationsByType(notifications, 'assigned').length).toBe(1);
  });
});

describe('sortNotificationsDesc', () => {
  it('sorts newest first', () => {
    const notifications = [
      makeNotification({ id: 'a', timestamp: '2026-01-01T00:00:00Z' }),
      makeNotification({ id: 'b', timestamp: '2026-06-15T00:00:00Z' }),
    ];
    expect(sortNotificationsDesc(notifications)[0].id).toBe('b');
  });
});

describe('notificationsForTask', () => {
  it('returns notifications for a task', () => {
    const notifications = [
      makeNotification({ taskId: 't1' }),
      makeNotification({ id: 'n2', taskId: 't2' }),
    ];
    expect(notificationsForTask(notifications, 't1').length).toBe(1);
  });
});

describe('notificationTypeLabel', () => {
  it('returns readable labels', () => {
    expect(notificationTypeLabel('assigned')).toBe('Assigned to you');
    expect(notificationTypeLabel('mention')).toBe('Mentioned you');
    expect(notificationTypeLabel('due_soon')).toBe('Due soon');
  });
});

describe('hasUnreadNotification', () => {
  it('returns true when unread exist', () => {
    expect(hasUnreadNotification([makeNotification({ read: false })])).toBe(true);
  });

  it('returns false when all read', () => {
    expect(hasUnreadNotification([makeNotification({ read: true })])).toBe(false);
  });
});
