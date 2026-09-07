import React from 'react';
import { render, screen } from '@testing-library/react';
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
    useFlowStore: () => flowStoreState,
  };
});

const baseTask: Task = {
  id: 't1',
  projectId: 'p1',
  workspaceId: '',
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

  it('does not paint the status colour straight onto its own tint', () => {
    // This used to assert `color: #ec4899`, which pinned the defect: the badge
    // sat on a 13% tint of that same colour, and #ec4899 on it measures 4.64:1
    // in dark and 3.13:1 in light (rtylr#1246). The foreground is now derived
    // by `onTint`.
    //
    // The derived value itself cannot be asserted here. happy-dom's
    // CSSStyleDeclaration silently discards any `color-mix()` assigned through
    // the style property, which is the path React uses, so both the tint and the
    // foreground are absent from the DOM in this environment (rtylr#1249). What
    // is checkable is that the raw colour is not the foreground, and
    // `onTint.test.ts` measures the derived value directly.
    render(<TaskCard task={{ ...baseTask, statusName: 'Legal review', statusColor: '#ec4899' }} />);

    const badge = screen.getByText('Legal review');
    expect(badge.style.color).not.toBe('#ec4899');
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
