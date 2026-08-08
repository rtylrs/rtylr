import { describe, expect, it, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { TaskRow } from './ProjectList';
import type { Task } from '../../shared/stores/flowStore';

const mockNavigate = vi.fn();

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
  useParams: () => ({}),
  Link: ({ children }: { children?: ReactNode }) => <>{children}</>,
}));

const task: Task = {
  id: 't1',
  projectId: 'p-web',
  workspaceId: 'w1',
  title: 'Ship the release',
  status: 'todo',
  priority: 'medium',
  assignee: 'Ada Lovelace',
  assigneeInitials: 'AL',
  assigneeId: 'u1',
  dueDate: '',
  subtaskCount: 0,
  projectName: 'Web',
  createdAt: '',
  updatedAt: '',
  identifier: 'FLOW-1',
};

function renderRow(): void {
  const queryClient = new QueryClient();
  render(
    <QueryClientProvider client={queryClient}>
      <TaskRow task={task} />
    </QueryClientProvider>
  );
}

describe('ProjectList TaskRow more menu', () => {
  it('opens a menu with open/copy link/delete actions when the more button is clicked', () => {
    renderRow();

    fireEvent.click(screen.getByRole('button', { name: /more options/i }));

    expect(screen.getByText('Open task')).toBeInTheDocument();
    expect(screen.getByText('Copy link')).toBeInTheDocument();
    expect(screen.getByText('Delete')).toBeInTheDocument();
  });

  it('navigates to the task detail page when "Open task" is clicked', () => {
    renderRow();

    fireEvent.click(screen.getByRole('button', { name: /more options/i }));
    fireEvent.click(screen.getByText('Open task'));

    expect(mockNavigate).toHaveBeenCalledWith({ to: '/task/t1' });
  });
});
