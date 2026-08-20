import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { toast } from 'sonner';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const myIssuesData = vi.hoisted(() => ({ data: [] as unknown[], isLoading: false }));
vi.mock('./useMyIssues', () => ({
  useMyIssues: () => myIssuesData,
}));

const createTaskMutate = vi.hoisted(() => vi.fn());
vi.mock('./useTasks', () => ({
  useCreateTask: () => ({ mutate: createTaskMutate, isPending: false }),
}));

const myTasksProjectData = vi.hoisted(() => ({
  data: { projectId: '9', workspaceId: '3' } as
    | { projectId: string; workspaceId: string }
    | undefined,
}));
vi.mock('./useMyTasksProject', () => ({
  useMyTasksProject: () => myTasksProjectData,
}));

const labelsData = vi.hoisted(() => ({
  data: [] as { id: number; name: string }[],
  isLoading: false,
}));
vi.mock('./useLabels', () => ({
  useLabels: () => labelsData,
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ selectedTaskIds: [], activeWorkspaceId: 'ws-1' }),
}));

vi.mock('../../shared/lib/currentUser', () => ({
  getCurrentUserId: () => 'u1',
}));

vi.mock('../../shared/components/TaskCard', () => ({
  default: ({ task }: { task: { title: string } }) => <div>{task.title}</div>,
}));

import MyIssues from './MyIssues';

beforeEach(() => {
  vi.clearAllMocks();
  myIssuesData.data = [];
  myIssuesData.isLoading = false;
  myTasksProjectData.data = { projectId: '9', workspaceId: '3' };
  labelsData.data = [];
  labelsData.isLoading = false;
});

describe('MyIssues QuickAddRow', () => {
  it('blocks submit and shows an actionable toast when the workspace has no labels', () => {
    labelsData.data = [];
    render(<MyIssues />);

    const input = screen.getByPlaceholderText('flow.myIssues.quickAddPlaceholder');
    fireEvent.change(input, { target: { value: 'Ship the thing' } });
    fireEvent.keyDown(input, { key: 'Enter' });

    expect(createTaskMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('flow.myIssues.quickAddNoLabelsReason');
  });

  it('disables the Add button with an explanatory title when the workspace has no labels', () => {
    labelsData.data = [];
    render(<MyIssues />);

    fireEvent.change(screen.getByPlaceholderText('flow.myIssues.quickAddPlaceholder'), {
      target: { value: 'Ship the thing' },
    });
    const button = screen.getByRole('button', { name: 'flow.myIssues.quickAddButton' });

    expect(button).toBeDisabled();
    expect(button).toHaveAttribute('title', 'flow.myIssues.quickAddNoLabelsReason');
  });

  it('leaves the Add button enabled while labels are still loading', () => {
    labelsData.data = [];
    labelsData.isLoading = true;
    render(<MyIssues />);

    fireEvent.change(screen.getByPlaceholderText('flow.myIssues.quickAddPlaceholder'), {
      target: { value: 'Ship the thing' },
    });
    const button = screen.getByRole('button', { name: 'flow.myIssues.quickAddButton' });

    expect(button).not.toBeDisabled();
  });

  it('sends the first available label id in the create payload', () => {
    labelsData.data = [
      { id: 42, name: 'Bug' },
      { id: 43, name: 'Feature' },
    ];
    render(<MyIssues />);

    fireEvent.change(screen.getByPlaceholderText('flow.myIssues.quickAddPlaceholder'), {
      target: { value: 'Ship the thing' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.myIssues.quickAddButton' }));

    expect(createTaskMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        workspaceId: '3',
        projectId: '9',
        title: 'Ship the thing',
        labels: ['42'],
      }),
      expect.anything()
    );
    expect(toast.error).not.toHaveBeenCalled();
  });

  it('blocks submit when the personal project has not resolved yet', () => {
    myTasksProjectData.data = undefined;
    labelsData.data = [{ id: 42, name: 'Bug' }];
    render(<MyIssues />);

    fireEvent.change(screen.getByPlaceholderText('flow.myIssues.quickAddPlaceholder'), {
      target: { value: 'Ship the thing' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.myIssues.quickAddButton' }));

    expect(createTaskMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('errors.generic');
  });
});

describe('MyIssues empty state', () => {
  it('shows the illustration, title and a call-to-action when there are no tasks', () => {
    myIssuesData.data = [];
    render(<MyIssues />);

    expect(screen.getByText('flow.myIssues.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.myIssues.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.myIssues.emptyCta' })).toBeInTheDocument();
  });

  it('focuses the quick-add input when the call-to-action is clicked', () => {
    myIssuesData.data = [];
    render(<MyIssues />);

    fireEvent.click(screen.getByRole('button', { name: 'flow.myIssues.emptyCta' }));

    expect(screen.getByPlaceholderText('flow.myIssues.quickAddPlaceholder')).toHaveFocus();
  });

  it('does not show the empty state when there are tasks', () => {
    myIssuesData.data = [{ id: '1', title: 'Ship the thing', status: 'todo' }];
    render(<MyIssues />);

    expect(screen.queryByText('flow.myIssues.emptyTitle')).not.toBeInTheDocument();
  });
});
