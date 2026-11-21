import React from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ApiTaskActivity, ApiTimeLog } from '../../../shared/services/flowApi';
import type { Comment } from '../../../shared/stores/flowStore';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    // Keys come back as-is, with any values appended, so a test can see both
    // which string was used and what was put into it.
    t: (key: string, values?: Record<string, unknown>) =>
      values ? `${key}(${Object.values(values).join('|')})` : key,
    i18n: { language: 'en' },
  }),
}));

const timeLogs = vi.hoisted(() => ({
  data: { timeLogs: [] as ApiTimeLog[] },
  isPending: false,
  isError: false,
  refetch: vi.fn(),
}));
vi.mock('../useTimeLogs', () => ({ useTimeLogs: () => timeLogs }));

import TaskActivityTabs, { buildFeed, historyEntries } from './TaskActivityTabs';

const comment = (id: string, body: string, createdAt: string): Comment => ({
  id,
  taskId: '1',
  authorId: '7',
  author: 'Abed',
  authorInitials: 'A',
  body,
  createdAt,
});
const activity = (
  id: number,
  action: string,
  newValue: string,
  createdAt: string
): ApiTaskActivity => ({
  id,
  taskId: 1,
  userId: 7,
  action,
  oldValue: '',
  newValue,
  createdAt,
});
const timeLog = (id: number, minutes: number, createdAt: string): ApiTimeLog => ({
  id,
  taskId: 1,
  userId: 31,
  durationMinutes: minutes,
  description: 'Pairing on the print view',
  loggedAt: createdAt,
  createdAt,
});

const comments = [comment('c1', 'First comment', '2026-11-18T09:00:00Z')];
const entries = [
  activity(2, 'assignee_changed', '31', '2026-11-18T10:00:00Z'),
  // The server logs every comment as activity too; History must skip these.
  activity(1, 'comment_added', 'First comment', '2026-11-18T09:00:00Z'),
];
const people = { '7': 'Abed', '31': 'Mohamad Homsi' };

const renderTabs = (props: Partial<React.ComponentProps<typeof TaskActivityTabs>> = {}) =>
  render(
    <TaskActivityTabs
      taskId="1"
      projectId="8"
      currentUserId="7"
      comments={comments}
      activity={entries}
      peopleNames={people}
      renderComments={(list) => (
        <div>
          {list.map((c) => (
            <p key={c.id}>comment: {c.body}</p>
          ))}
        </div>
      )}
      composer={<div>composer</div>}
      {...props}
    />
  );

const panel = () => screen.getByRole('tabpanel');

beforeEach(() => {
  window.localStorage.clear();
  timeLogs.data = { timeLogs: [timeLog(5, 90, '2026-11-18T11:00:00Z')] };
  timeLogs.isPending = false;
  timeLogs.isError = false;
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('activity feed helpers', () => {
  it('drops comment rows from history, since Comments already shows them', () => {
    expect(historyEntries(entries).map((e) => e.action)).toEqual(['assignee_changed']);
  });

  it('orders comments, changes and time entries oldest first', () => {
    const feed = buildFeed(comments, entries, timeLogs.data.timeLogs);
    expect(feed.map((item) => item.key)).toEqual(['comment-c1', 'history-2', 'time-5']);
  });

  it('places a time entry by when it was recorded, not the day it was for', () => {
    const late = { ...timeLog(6, 30, '2026-11-18T12:00:00Z'), loggedAt: '2026-11-01' };
    const feed = buildFeed(comments, [], [late]);
    expect(feed.map((item) => item.key)).toEqual(['comment-c1', 'time-6']);
  });
});

describe('TaskActivityTabs', () => {
  it('opens on Comments with the comment box, as the panel did before', () => {
    renderTabs();
    expect(screen.getByRole('tab', { name: /flow\.taskDetail\.comments/ })).toHaveAttribute(
      'aria-selected',
      'true'
    );
    expect(within(panel()).getByText('comment: First comment')).toBeInTheDocument();
    expect(within(panel()).getByText('composer')).toBeInTheDocument();
  });

  it('shows changes without comments under History, with names instead of ids', () => {
    renderTabs();
    fireEvent.click(screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.history' }));
    expect(
      within(panel()).getByText(/flow\.taskDetail\.assigneeChanged\(Mohamad Homsi\)/)
    ).toBeInTheDocument();
    expect(within(panel()).queryByText(/First comment/)).not.toBeInTheDocument();
    expect(within(panel()).queryByText('composer')).not.toBeInTheDocument();
  });

  it('lists the task time entries under Time', () => {
    renderTabs();
    const time = screen.getByRole('tab', { name: /flow\.taskDetail\.activityTabs\.time/ });
    expect(time).toHaveTextContent('1');
    fireEvent.click(time);
    expect(
      within(panel()).getByText(/flow\.taskDetail\.activityTabs\.timeLogged\(Mohamad Homsi\|1:30\)/)
    ).toBeInTheDocument();
    expect(within(panel()).getByText('Pairing on the print view')).toBeInTheDocument();
  });

  it('says so when no time has been logged', () => {
    timeLogs.data = { timeLogs: [] };
    renderTabs();
    fireEvent.click(screen.getByRole('tab', { name: /flow\.taskDetail\.activityTabs\.time/ }));
    expect(within(panel()).getByText('flow.time.empty')).toBeInTheDocument();
  });

  it('puts everything in one timeline under All, ending at the comment box', () => {
    renderTabs();
    fireEvent.click(screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.all' }));
    const text = panel().textContent ?? '';
    const order = ['comment: First comment', 'assigneeChanged', 'timeLogged', 'composer'].map(
      (part) => text.indexOf(part)
    );
    expect(order.every((at) => at >= 0)).toBe(true);
    expect([...order].sort((a, b) => a - b)).toEqual(order);
  });

  it('moves between tabs with the arrow, Home and End keys', () => {
    renderTabs();
    const tablist = screen.getByRole('tablist');
    const comments = screen.getByRole('tab', { name: /flow\.taskDetail\.comments/ });
    comments.focus();
    fireEvent.keyDown(tablist, { key: 'ArrowRight' });
    const history = screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.history' });
    expect(history).toHaveAttribute('aria-selected', 'true');
    expect(history).toHaveFocus();
    expect(history).toHaveAttribute('tabindex', '0');
    expect(comments).toHaveAttribute('tabindex', '-1');
    fireEvent.keyDown(tablist, { key: 'Home' });
    expect(screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.all' })).toHaveFocus();
    fireEvent.keyDown(tablist, { key: 'ArrowLeft' });
    expect(screen.getByRole('tab', { name: /activityTabs\.time/ })).toHaveFocus();
    fireEvent.keyDown(tablist, { key: 'End' });
    expect(screen.getByRole('tab', { name: /activityTabs\.time/ })).toHaveAttribute(
      'aria-selected',
      'true'
    );
  });

  it('moves on from the focused tab even when it is not the selected one', () => {
    renderTabs();
    screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.history' }).focus();
    fireEvent.keyDown(screen.getByRole('tablist'), { key: 'ArrowRight' });
    expect(screen.getByRole('tab', { name: /activityTabs\.time/ })).toHaveAttribute(
      'aria-selected',
      'true'
    );
  });

  it('remembers the last tab per person and project', () => {
    const first = renderTabs();
    fireEvent.click(screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.history' }));
    first.unmount();

    renderTabs();
    expect(
      screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.history' })
    ).toHaveAttribute('aria-selected', 'true');
  });

  it('falls back to Comments when storage is unavailable', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked');
    });
    renderTabs();
    expect(screen.getByRole('tab', { name: /flow\.taskDetail\.comments/ })).toHaveAttribute(
      'aria-selected',
      'true'
    );
    fireEvent.click(screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.history' }));
    expect(
      screen.getByRole('tab', { name: 'flow.taskDetail.activityTabs.history' })
    ).toHaveAttribute('aria-selected', 'true');
  });
});
