import { describe, it, expect } from 'vitest';
import {
  severityLabel,
  isUrgent,
  unreadNotifications,
  unreadCount,
  sortByDate,
  badgeLabel,
  isChannelEnabled,
} from './notificationUtils';
import type { Notification } from './notificationUtils';

const makeNotifications = (): Notification[] => [
  {
    id: '1',
    title: 'A',
    body: '',
    severity: 'info',
    read: false,
    createdAt: '2025-01-01T10:00:00Z',
  },
  {
    id: '2',
    title: 'B',
    body: '',
    severity: 'error',
    read: true,
    createdAt: '2025-01-03T08:00:00Z',
  },
  {
    id: '3',
    title: 'C',
    body: '',
    severity: 'warning',
    read: false,
    createdAt: '2025-01-02T12:00:00Z',
  },
];

describe('severityLabel', () => {
  it('returns "Info" for info', () => {
    expect(severityLabel('info')).toBe('Info');
  });

  it('returns "Success" for success', () => {
    expect(severityLabel('success')).toBe('Success');
  });

  it('returns "Warning" for warning', () => {
    expect(severityLabel('warning')).toBe('Warning');
  });

  it('returns "Error" for error', () => {
    expect(severityLabel('error')).toBe('Error');
  });
});

describe('isUrgent', () => {
  it('returns true for error', () => {
    expect(isUrgent('error')).toBe(true);
  });

  it('returns true for warning', () => {
    expect(isUrgent('warning')).toBe(true);
  });

  it('returns false for info', () => {
    expect(isUrgent('info')).toBe(false);
  });

  it('returns false for success', () => {
    expect(isUrgent('success')).toBe(false);
  });
});

describe('unreadNotifications', () => {
  it('returns only unread notifications', () => {
    const result = unreadNotifications(makeNotifications());
    expect(result.every((n) => !n.read)).toBe(true);
  });

  it('excludes read notifications', () => {
    const result = unreadNotifications(makeNotifications());
    expect(result.find((n) => n.id === '2')).toBeUndefined();
  });
});

describe('unreadCount', () => {
  it('counts unread notifications', () => {
    expect(unreadCount(makeNotifications())).toBe(2);
  });

  it('returns 0 for empty array', () => {
    expect(unreadCount([])).toBe(0);
  });
});

describe('sortByDate', () => {
  it('sorts newest first', () => {
    const sorted = sortByDate(makeNotifications());
    expect(sorted[0]!.id).toBe('2');
    expect(sorted[1]!.id).toBe('3');
    expect(sorted[2]!.id).toBe('1');
  });

  it('does not mutate the original array', () => {
    const notifications = makeNotifications();
    sortByDate(notifications);
    expect(notifications[0]!.id).toBe('1');
  });
});

describe('badgeLabel', () => {
  it('returns empty string for 0', () => {
    expect(badgeLabel(0)).toBe('');
  });

  it('returns the count as a string for small counts', () => {
    expect(badgeLabel(5)).toBe('5');
  });

  it('returns "9+" when count exceeds maxDisplay', () => {
    expect(badgeLabel(12)).toBe('9+');
  });

  it('uses custom maxDisplay', () => {
    expect(badgeLabel(25, 20)).toBe('20+');
  });
});

describe('isChannelEnabled', () => {
  it('returns true when channel is in the enabled list', () => {
    expect(isChannelEnabled('email', ['email', 'push'])).toBe(true);
  });

  it('returns false when channel is not in the list', () => {
    expect(isChannelEnabled('sms', ['email', 'push'])).toBe(false);
  });

  it('returns false for empty enabled list', () => {
    expect(isChannelEnabled('email', [])).toBe(false);
  });
});
