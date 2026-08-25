import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const myIssuesData = vi.hoisted(() => ({ data: [] as unknown[], isLoading: false }));
vi.mock('./useMyIssues', () => ({
  useMyIssues: () => myIssuesData,
}));

const setQuickCreateOpen = vi.hoisted(() => vi.fn());

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ selectedTaskIds: [], setQuickCreateOpen }),
}));

vi.mock('../../shared/components/TaskCard', () => ({
  default: ({ task }: { task: { title: string } }) => <div>{task.title}</div>,
}));

import MyIssues from './MyIssues';

beforeEach(() => {
  vi.clearAllMocks();
  myIssuesData.data = [];
  myIssuesData.isLoading = false;
});

describe('MyIssues empty state', () => {
  it('shows the normal kanban columns and board-level CTA when there are no tasks', () => {
    myIssuesData.data = [];
    render(<MyIssues />);

    expect(screen.getByText('flow.taskDetail.backlog')).toBeInTheDocument();
    expect(screen.getByText('flow.taskDetail.todo')).toBeInTheDocument();
    expect(screen.getByText('flow.taskDetail.inProgress')).toBeInTheDocument();
    expect(screen.getByText('flow.taskDetail.inReview')).toBeInTheDocument();
    expect(screen.getByText('flow.taskDetail.done')).toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.board.addTaskLabel' })).toBeInTheDocument();
  });

  it('opens the full task form from the board CTA', () => {
    myIssuesData.data = [];
    render(<MyIssues />);

    fireEvent.click(screen.getByRole('button', { name: 'flow.board.addTaskLabel' }));

    expect(setQuickCreateOpen).toHaveBeenCalledWith(true);
  });

  it('does not show the board-level welcome when there are tasks', () => {
    myIssuesData.data = [{ id: '1', title: 'Ship the thing', status: 'todo' }];
    render(<MyIssues />);

    expect(screen.queryByText('flow.myIssues.emptyTitle')).not.toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.board.addTaskLabel' })
    ).not.toBeInTheDocument();
  });

  it('does not render the old inline quick-add field', () => {
    render(<MyIssues />);

    expect(
      screen.queryByPlaceholderText('flow.myIssues.quickAddPlaceholder')
    ).not.toBeInTheDocument();
  });
});
