// @vitest-environment happy-dom
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import type { Task } from '../stores/flowStore';

vi.mock('react-i18next', async (importOriginal) => ({
  ...(await importOriginal<typeof import('react-i18next')>()),
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

vi.mock('@tanstack/react-router', () => ({ useNavigate: () => vi.fn() }));

// TaskRow renders its own chip markup rather than using LabelChip, so it is
// covered separately from TaskCard even though the defect was shared.
const labelsQuery = vi.hoisted(() => ({
  data: [] as { id: string | number; name: string; color: string }[],
}));
vi.mock('../../features/tasks/useLabels', () => ({ useLabels: () => labelsQuery }));

// labels stays [] because setLabels has no call site in the app: keeping the
// store empty here is what makes these tests meaningful.
const flowStoreState = vi.hoisted(() => ({
  labels: [],
  projects: [],
  toggleSelectTask: vi.fn(),
}));
vi.mock('../stores/flowStore', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../stores/flowStore')>()),
  useFlowStore: () => flowStoreState,
}));

import { TaskRow } from './TaskRow';

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
  projectName: 'Project',
  createdAt: '2026-01-01',
  updatedAt: '2026-01-01',
  identifier: 'FLOW-1',
  labelIds: [],
} as unknown as Task;

describe('TaskRow labels (#1039)', () => {
  beforeEach(() => {
    labelsQuery.data = [
      { id: 994343, name: 'qa-temp', color: '#ef4444' },
      { id: 994344, name: 'blocked', color: '#22c55e' },
    ];
  });

  it('renders only the labels the task carries', () => {
    render(<TaskRow task={{ ...baseTask, labelIds: ['994344'] }} />);

    expect(screen.getByText('blocked')).toBeInTheDocument();
    expect(screen.queryByText('qa-temp')).not.toBeInTheDocument();
  });

  it('renders nothing when the task has no labels', () => {
    render(<TaskRow task={{ ...baseTask, labelIds: [] }} />);

    expect(screen.queryByText('blocked')).not.toBeInTheDocument();
  });

  it('caps the visible chips at three and counts the rest', () => {
    labelsQuery.data = [1, 2, 3, 4, 5].map((n) => ({
      id: n,
      name: `label-${n}`,
      color: '#ef4444',
    }));

    render(<TaskRow task={{ ...baseTask, labelIds: ['1', '2', '3', '4', '5'] }} />);

    expect(screen.getByText('label-1')).toBeInTheDocument();
    expect(screen.getByText('label-3')).toBeInTheDocument();
    expect(screen.queryByText('label-4')).not.toBeInTheDocument();
    expect(screen.getByText('+2')).toBeInTheDocument();
  });
});

describe('TaskRow due date colour (#1077)', () => {
  const at = (iso: string): void => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(iso));
  };

  afterEach(() => {
    vi.useRealTimers();
  });

  const dateEl = (dueDate: string): HTMLElement => {
    render(<TaskRow task={{ ...baseTask, dueDate, status: 'todo' }} />);
    // Both the due date and the task identifier carry a 64px column, and the
    // identifier comes first in the DOM, so the right-aligned one is the date.
    const el = document.querySelector('span.w-\\[64px\\].text-right');
    expect(el, 'the due date cell').not.toBeNull();
    return el as HTMLElement;
  };

  it('paints a late date with the danger tone', () => {
    at('2026-05-02T12:00:00.000Z');
    expect(dateEl('2026-05-01').style.color).toBe('var(--flow-danger-fg)');
  });

  it('paints a date due today with the warn tone, not danger', () => {
    // The bug this test exists for: `new Date('2026-05-02') < new Date()` is
    // true from 00:00 UTC, so a task due today was painted red for the whole
    // working day at UTC+3.
    at('2026-05-02T12:00:00.000Z');
    expect(dateEl('2026-05-02').style.color).toBe('var(--flow-warn-fg)');
  });

  it('paints a future date with the muted tone, never the dim one', () => {
    // --flow-text-dim measures 2.35:1 in both modes and globals.css says it
    // never carries text. It was the colour of every non-overdue due date.
    at('2026-05-02T12:00:00.000Z');
    const el = dateEl('2026-06-01');
    expect(el.style.color).toBe('var(--flow-text-muted)');
    expect(el.style.color).not.toBe('var(--flow-text-dim)');
  });
});
