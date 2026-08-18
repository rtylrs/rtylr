import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { toast } from 'sonner';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const flowStoreState = vi.hoisted(() => ({
  quickCreateOpen: true,
  setQuickCreateOpen: vi.fn(),
  projects: [{ id: 'p1', name: 'Project One', workspaceId: 'ws-1', color: '#fff' }],
  activeWorkspaceId: 'ws-1',
  sections: [],
}));
vi.mock('../stores/flowStore', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../stores/flowStore')>();
  return {
    ...actual,
    useFlowStore: () => flowStoreState,
  };
});

vi.mock('../stores/userStore', () => ({
  useUserStore: (selector: (s: { user: unknown }) => unknown) =>
    selector({ user: { id: 'u1', name: 'Ada Lovelace', email: 'ada@example.com' } }),
}));

vi.mock('../lib/currentUser', () => ({
  getCurrentUserId: () => 'u1',
}));

const createTaskMutateAsync = vi.hoisted(() => vi.fn());
vi.mock('../../features/tasks/useTasks', () => ({
  useCreateTask: () => ({ mutateAsync: createTaskMutateAsync, isPending: false }),
}));

const labelsData = vi.hoisted(() => ({ data: [] as { id: number; name: string }[] }));
vi.mock('../../features/tasks/useLabels', () => ({
  useLabels: () => labelsData,
}));

import QuickCreate from './QuickCreate';

beforeEach(() => {
  vi.clearAllMocks();
  flowStoreState.quickCreateOpen = true;
  labelsData.data = [];
});

function fillTitle(): void {
  fireEvent.change(screen.getByPlaceholderText('flow.quickCreate.titlePlaceholder'), {
    target: { value: 'Ship the thing' },
  });
}

describe('QuickCreate label requirement', () => {
  it('renders the workspace labels in the label selector', () => {
    labelsData.data = [
      { id: 1, name: 'Bug' },
      { id: 2, name: 'Feature' },
    ];
    render(<QuickCreate />);

    const select = screen.getByLabelText('common.settings.labels') as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.textContent)).toEqual(['Bug', 'Feature']);
  });

  it('blocks submit and shows an error toast when the workspace has no labels', async () => {
    labelsData.data = [];
    render(<QuickCreate />);

    fillTitle();
    fireEvent.click(screen.getByRole('button', { name: 'flow.quickCreate.createTask' }));

    expect(createTaskMutateAsync).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('errors.generic');
  });

  it('sends the selected label id in the create payload', async () => {
    labelsData.data = [
      { id: 7, name: 'Bug' },
      { id: 8, name: 'Feature' },
    ];
    createTaskMutateAsync.mockResolvedValue({
      task: { id: 't-1', identifier: 'FLO-1', title: 'x' },
    });
    render(<QuickCreate />);

    fillTitle();
    fireEvent.change(screen.getByLabelText('common.settings.labels'), { target: { value: '8' } });
    fireEvent.click(screen.getByRole('button', { name: 'flow.quickCreate.createTask' }));

    expect(createTaskMutateAsync).toHaveBeenCalledWith(expect.objectContaining({ labels: ['8'] }));
  });
});
