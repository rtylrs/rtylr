// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import type { Task } from '../stores/flowStore';
import type { ApiTaskRelationGroup } from '../services/flowApi';

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({
    t: (key: string, opts?: { count?: number }) =>
      opts?.count === undefined ? key : `${key}:${opts.count}`,
    i18n: { language: 'en-US' },
  }),
}));
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => vi.fn() }));
vi.mock('../../features/tasks/useLabels', () => ({ useLabels: () => ({ data: [] }) }));
vi.mock('../stores/flowStore', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../stores/flowStore')>()),
  useFlowStore: (selector?: (state: object) => unknown) => {
    const state = { labels: [], projects: [], toggleSelectTask: vi.fn() };
    return selector ? selector(state) : state;
  },
}));

const relations = vi.hoisted(() => ({
  enabledFor: [] as Array<[string | undefined, boolean | undefined]>,
  groups: [] as ApiTaskRelationGroup[],
}));
vi.mock('../../features/tasks/useTaskRelations', () => ({
  useTaskRelations: (taskId: string | undefined, enabled?: boolean) => {
    relations.enabledFor.push([taskId, enabled]);
    return enabled
      ? { data: relations.groups, isPending: false }
      : { data: undefined, isPending: true };
  },
}));

import { TaskRow } from './TaskRow';
import { TaskCard } from './TaskCard';

const baseTask = {
  id: 't1',
  projectId: 'p1',
  title: 'Ship the thing',
  status: 'todo',
  priority: 'medium',
  assignee: '',
  assigneeInitials: '',
  assigneeId: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Project',
  createdAt: '2026-01-01',
  updatedAt: '2026-01-01',
  identifier: 'T-1',
  labelIds: [],
} as unknown as Task;

const blocker = (id: string, title: string, statusBucket: 'active' | 'done') => ({
  relationId: `r${id}`,
  taskId: id,
  identifier: `T-${id}`,
  title,
  projectId: 'p1',
  projectName: 'Project',
  status: 'todo',
  statusName: '',
  statusColor: '',
  statusBucket,
  assigneeName: '',
});

describe('BlockedBadge', () => {
  beforeEach(() => {
    relations.enabledFor = [];
    relations.groups = [
      {
        type: 'blocks',
        direction: 'inward',
        relation: 'blocked_by',
        tasks: [blocker('7', 'Design sign-off', 'active'), blocker('8', 'Old blocker', 'done')],
      },
    ];
  });

  it('is absent while nothing open blocks the task', () => {
    render(<TaskRow task={{ ...baseTask, blockerCount: 0 }} />);
    render(<TaskCard task={baseTask} />);
    expect(screen.queryByTestId('blocked-badge')).not.toBeInTheDocument();
  });

  it.each([
    ['list row', (task: Task) => <TaskRow task={task} />],
    ['board card', (task: Task) => <TaskCard task={task} />],
  ])('shows on a %s and lists only the open blockers on hover', (_where, view) => {
    render(view({ ...baseTask, blockerCount: 1 }));
    const badge = screen.getByLabelText('flow.blockedBadge.label:1');
    expect(badge).toHaveTextContent('flow.blockedBadge.text');
    // Nothing is fetched for a badge nobody has pointed at.
    expect(relations.enabledFor.every(([, enabled]) => enabled === false)).toBe(true);

    fireEvent.mouseEnter(screen.getByTestId('blocked-badge'));
    fireEvent.mouseEnter(badge.parentElement!);
    expect(relations.enabledFor.at(-1)).toEqual(['t1', true]);
    const tooltip = screen.getByRole('tooltip');
    expect(tooltip).toHaveTextContent('flow.blockedBadge.tooltipTitle:1');
    expect(tooltip).toHaveTextContent('T-7 · Design sign-off');
    expect(tooltip).not.toHaveTextContent('Old blocker');
  });
});
