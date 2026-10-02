import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../testing/i18n';
import { TaskCard } from './TaskCard';
import type { Task } from '../stores/flowStore';

const openTask = vi.hoisted(() => vi.fn());
vi.mock('../lib/useOpenTask', () => ({ useOpenTask: () => openTask }));
vi.mock('../../features/tasks/useLabels', () => ({ useLabels: () => ({ data: [] }) }));
vi.mock('../stores/flowStore', async (load) => ({
  ...(await load<typeof import('../stores/flowStore')>()),
  useFlowStore: () => ({ projects: [] }),
}));
beforeEach(async () => {
  await initTestI18n();
});
const task: Task = {
  id: '1',
  identifier: 'T-1',
  title: 'Keyboard verification',
  projectId: '2',
  status: 'todo',
  priority: 'no_priority',
  assignee: '',
  assigneeInitials: '',
  assigneeId: '',
  dueDate: '',
  subtaskCount: 0,
  projectName: '',
  createdAt: '',
  updatedAt: '',
};

describe('board task keyboard navigation', () => {
  it.each(['Enter', ' '])('opens a focused task with %s', (key) => {
    openTask.mockClear();
    render(<TaskCard task={task} />);
    fireEvent.keyDown(
      screen.getByRole('button', { name: 'Open task T-1: Keyboard verification' }),
      { key }
    );
    expect(openTask).toHaveBeenCalledWith('1');
  });

  it('lets a nested control handle its own keyboard activation', () => {
    openTask.mockClear();
    render(<TaskCard task={task} controls={<button>Move task</button>} />);
    fireEvent.keyDown(screen.getByRole('button', { name: 'Move task' }), { key: 'Enter' });
    expect(openTask).not.toHaveBeenCalled();
  });
});
