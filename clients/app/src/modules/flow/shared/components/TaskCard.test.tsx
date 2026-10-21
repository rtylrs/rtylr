// @vitest-environment jsdom
//
// jsdom rather than the suite default. happy-dom silently discards any
// color-mix() assigned through the style property, which is the path React uses,
// so the tint and the foreground derived from it are simply absent from the DOM
// there and no assertion can see them (rtylr#1249). jsdom keeps both.
import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, it, expect, vi } from 'vitest';
import { TaskCard } from './TaskCard';
import type { Task } from '../stores/flowStore';

vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return {
    ...actual,
    useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
  };
});

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

const openTask = vi.hoisted(() => vi.fn());
vi.mock('../lib/useOpenTask', () => ({ useOpenTask: () => openTask }));

// The label catalogue comes from the query now, not the store: `setLabels` has
// no call site in the app, so the store's array is permanently empty in
// production and keeping it that way here stops a test passing on data the app
// never has.
const labelsQuery = vi.hoisted(() => ({
  data: [] as { id: string | number; name: string; color: string }[],
}));
vi.mock('../../features/tasks/useLabels', () => ({
  useLabels: () => labelsQuery,
}));

const flowStoreState = vi.hoisted(() => ({
  labels: [],
  comments: [],
  attachments: [],
  subtasks: [],
  projects: [],
}));
vi.mock('../stores/flowStore', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../stores/flowStore')>();
  return {
    ...actual,
    useFlowStore: (selector?: (state: object) => unknown) => {
      const state = flowStoreState;
      return selector ? selector(state) : state;
    },
  };
});

const baseTask: Task = {
  id: 't1',
  projectId: 'p1',
  title: 'Ship the thing',
  status: 'in_progress',
  priority: 'medium',
  assignee: '',
  assigneeInitials: '',
  assigneeId: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: '',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  identifier: 'FLOW-1',
};

describe('TaskCard custom status badge (F-A)', () => {
  it('renders the custom status name and carries its colour', () => {
    render(<TaskCard task={{ ...baseTask, statusName: 'Legal review', statusColor: '#ec4899' }} />);

    const badge = screen.getByText('Legal review');
    expect(badge).toBeInTheDocument();
    // The status colour reaches the badge as `--tone`, which both the tint and
    // the foreground read back.
    expect(badge.style.getPropertyValue('--tone')).toBe('#ec4899');
  });

  it('tints the badge and derives its foreground rather than reusing the colour', () => {
    // This used to assert `color: #ec4899`, which pinned the defect: the badge
    // sits on a 13% tint of that same colour, and #ec4899 on it measures 4.64:1
    // in dark and 3.13:1 in light (rtylr#1246).
    //
    // Both declarations are only visible because this file runs under jsdom; see
    // the note at the top.
    render(<TaskCard task={{ ...baseTask, statusName: 'Legal review', statusColor: '#ec4899' }} />);

    const badge = screen.getByText('Legal review');
    expect(badge.style.background).toBe('color-mix(in srgb, var(--tone) 13%, transparent)');
    expect(badge.style.color).toBe('color-mix(in srgb, var(--tone) 50%, var(--flow-text))');
    // The specific thing that was wrong: the raw colour as the foreground.
    expect(badge.style.color).not.toContain('#ec4899');
  });

  it('renders no badge when the task has no custom status', () => {
    render(<TaskCard task={baseTask} />);

    expect(screen.queryByText('Legal review')).not.toBeInTheDocument();
  });
});

describe('TaskCard labels (#1039)', () => {
  beforeEach(() => {
    // Numeric ids on purpose: ApiLabel declares a number, and that is what the
    // old store-backed filter could never match against the task's strings.
    labelsQuery.data = [
      { id: 994343, name: 'qa-temp', color: '#ef4444' },
      { id: 994344, name: 'blocked', color: '#22c55e' },
    ];
  });

  it('renders a chip for each label the task carries', () => {
    render(<TaskCard task={{ ...baseTask, labelIds: ['994343'] }} />);

    expect(screen.getByText('qa-temp')).toBeInTheDocument();
    expect(screen.queryByText('blocked')).not.toBeInTheDocument();
  });

  it('renders nothing when the task has no labels', () => {
    render(<TaskCard task={{ ...baseTask, labelIds: [] }} />);

    expect(screen.queryByText('qa-temp')).not.toBeInTheDocument();
    expect(screen.queryByText('blocked')).not.toBeInTheDocument();
  });

  it('renders no chip when the catalogue has not loaded', () => {
    labelsQuery.data = [];

    render(<TaskCard task={{ ...baseTask, labelIds: ['994343'] }} />);

    expect(screen.queryByText('qa-temp')).not.toBeInTheDocument();
  });
});

describe('TaskCard persisted activity', () => {
  it('renders remote comment, attachment and completion counts with an empty local store', () => {
    render(
      <TaskCard
        task={{
          ...baseTask,
          commentCount: 3,
          attachmentCount: 2,
          subtaskCount: 4,
          completedSubtasks: 1,
        }}
      />
    );
    expect(screen.getByText('3')).toBeInTheDocument();
    expect(screen.getByText('2')).toBeInTheDocument();
    expect(screen.getByText('1/4')).toBeInTheDocument();
  });
});

describe('TaskCard board dragging', () => {
  it('starts a drag from the card surface and does not open the task after a drag', () => {
    const onDragPointerDown = vi.fn();
    const onDragPointerEnd = vi.fn().mockReturnValueOnce(false).mockReturnValueOnce(true);
    render(
      <TaskCard
        task={baseTask}
        onDragPointerDown={onDragPointerDown}
        onDragPointerEnd={onDragPointerEnd}
      />
    );

    const card = screen.getByRole('button', { name: 'flowCommon.taskRow.openTask' });
    fireEvent.pointerDown(card, { button: 0 });
    fireEvent.pointerUp(card);
    fireEvent.click(card);
    expect(openTask).toHaveBeenCalledWith(baseTask.id);

    openTask.mockClear();
    fireEvent.pointerDown(card, { button: 0 });
    fireEvent.pointerUp(card);
    fireEvent.click(card);

    expect(onDragPointerDown).toHaveBeenCalledTimes(2);
    expect(onDragPointerDown).toHaveBeenLastCalledWith(expect.anything(), baseTask.id);
    expect(onDragPointerEnd).toHaveBeenCalledTimes(2);
    expect(openTask).not.toHaveBeenCalled();
  });
});
