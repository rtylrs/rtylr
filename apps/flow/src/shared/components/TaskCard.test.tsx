import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
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
  it('renders the custom status name and color when the task has one', () => {
    render(<TaskCard task={{ ...baseTask, statusName: 'Legal review', statusColor: '#ec4899' }} />);

    const badge = screen.getByText('Legal review');
    expect(badge).toBeInTheDocument();
    expect(badge).toHaveStyle({ color: '#ec4899' });
  });

  it('renders no badge when the task has no custom status', () => {
    render(<TaskCard task={baseTask} />);

    expect(screen.queryByText('Legal review')).not.toBeInTheDocument();
  });
});
