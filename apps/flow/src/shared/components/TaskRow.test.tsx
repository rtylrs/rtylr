// @vitest-environment happy-dom
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
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
