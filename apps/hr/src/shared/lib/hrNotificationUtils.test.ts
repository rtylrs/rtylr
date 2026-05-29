import { describe, it, expect } from 'vitest';
import {
  notificationTypeLabel,
  notificationPriorityLabel,
  notificationPriorityColor,
  isActionRequired,
  countUnread,
  filterByPriority,
  sortByPriorityDesc,
} from './hrNotificationUtils';

describe('notificationTypeLabel', () => {
  it('returns Leave Request for leave_request', () => {
    expect(notificationTypeLabel('leave_request')).toBe('Leave Request');
  });
  it('returns Work Anniversary for anniversary', () => {
    expect(notificationTypeLabel('anniversary')).toBe('Work Anniversary');
  });
  it('returns Payroll Processed for payroll_processed', () => {
    expect(notificationTypeLabel('payroll_processed')).toBe('Payroll Processed');
  });
});

describe('notificationPriorityLabel', () => {
  it('returns Urgent for urgent', () => {
    expect(notificationPriorityLabel('urgent')).toBe('Urgent');
  });
  it('returns Low for low', () => {
    expect(notificationPriorityLabel('low')).toBe('Low');
  });
});

describe('notificationPriorityColor', () => {
  it('returns red for urgent', () => {
    expect(notificationPriorityColor('urgent')).toBe('red');
  });
  it('returns blue for medium', () => {
    expect(notificationPriorityColor('medium')).toBe('blue');
  });
  it('returns gray for low', () => {
    expect(notificationPriorityColor('low')).toBe('gray');
  });
});

describe('isActionRequired', () => {
  it('returns true for leave_request', () => {
    expect(isActionRequired('leave_request')).toBe(true);
  });
  it('returns true for review_due', () => {
    expect(isActionRequired('review_due')).toBe(true);
  });
  it('returns false for birthday', () => {
    expect(isActionRequired('birthday')).toBe(false);
  });
  it('returns false for system', () => {
    expect(isActionRequired('system')).toBe(false);
  });
});

describe('countUnread', () => {
  it('counts unread notifications', () => {
    const notifications = [{ read: false }, { read: true }, { read: false }];
    expect(countUnread(notifications)).toBe(2);
  });
  it('returns 0 when all read', () => {
    expect(countUnread([{ read: true }, { read: true }])).toBe(0);
  });
});

describe('filterByPriority', () => {
  it('filters by priority', () => {
    const notifications = [
      { priority: 'high' as const, id: 1 },
      { priority: 'low' as const, id: 2 },
      { priority: 'high' as const, id: 3 },
    ];
    const result = filterByPriority(notifications, 'high');
    expect(result).toHaveLength(2);
    expect(result.every((n) => n.priority === 'high')).toBe(true);
  });
});

describe('sortByPriorityDesc', () => {
  it('sorts urgent first', () => {
    const notifications = [
      { priority: 'low' as const },
      { priority: 'urgent' as const },
      { priority: 'medium' as const },
    ];
    const sorted = sortByPriorityDesc(notifications);
    expect(sorted[0].priority).toBe('urgent');
    expect(sorted[sorted.length - 1].priority).toBe('low');
  });
  it('does not mutate original array', () => {
    const original = [{ priority: 'low' as const }, { priority: 'urgent' as const }];
    sortByPriorityDesc(original);
    expect(original[0].priority).toBe('low');
  });
});
