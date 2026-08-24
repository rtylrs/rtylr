import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { toast } from 'sonner';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
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

// Ruling M1: quick-add needs only a title. Labels are optional everywhere and the
// silent first-label stamping is gone.
//
// The suite this replaces was called "QuickCreate label requirement" and asserted
// the opposite: that submit is blocked and an error toast shown when the business
// has no labels. That was the behaviour M1 removes, so the test was pinning a
// decision the ledger had already reversed.
describe('QuickCreate labels are optional', () => {
  it('offers an explicit no-label choice rather than preselecting one', () => {
    labelsData.data = [
      { id: 1, name: 'Bug' },
      { id: 2, name: 'Feature' },
    ];
    render(<QuickCreate />);

    const select = screen.getByLabelText('common.settings.labels') as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.textContent)).toEqual([
      'flow.quickCreate.noLabel',
      'Bug',
      'Feature',
    ]);
    // Nothing is stamped for the user: the empty choice is what is selected.
    expect(select.value).toBe('');
  });

  // This is the case that was actually broken. The picker is disabled when there
  // are no labels, so the user cannot choose one - and the old code then refused
  // to submit for not having one. A business with no labels could not quick-create
  // a task at all, and the only feedback was a generic error.
  it('creates a task when the business has no labels at all', async () => {
    labelsData.data = [];
    createTaskMutateAsync.mockResolvedValue({
      task: { id: 't-1', identifier: 'FLO-1', title: 'x' },
    });
    render(<QuickCreate />);

    fillTitle();
    fireEvent.click(screen.getByRole('button', { name: 'flow.quickCreate.createTask' }));

    expect(createTaskMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ labels: undefined })
    );
    expect(toast.error).not.toHaveBeenCalled();
  });

  it('creates a task without a label when labels exist but none is chosen', async () => {
    labelsData.data = [{ id: 7, name: 'Bug' }];
    createTaskMutateAsync.mockResolvedValue({
      task: { id: 't-2', identifier: 'FLO-2', title: 'x' },
    });
    render(<QuickCreate />);

    fillTitle();
    fireEvent.click(screen.getByRole('button', { name: 'flow.quickCreate.createTask' }));

    expect(createTaskMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ labels: undefined })
    );
  });

  it('sends the selected label id when one is chosen', async () => {
    labelsData.data = [
      { id: 7, name: 'Bug' },
      { id: 8, name: 'Feature' },
    ];
    createTaskMutateAsync.mockResolvedValue({
      task: { id: 't-3', identifier: 'FLO-3', title: 'x' },
    });
    render(<QuickCreate />);

    fillTitle();
    fireEvent.change(screen.getByLabelText('common.settings.labels'), { target: { value: '8' } });
    fireEvent.click(screen.getByRole('button', { name: 'flow.quickCreate.createTask' }));

    expect(createTaskMutateAsync).toHaveBeenCalledWith(expect.objectContaining({ labels: ['8'] }));
  });

  // A title is still required: M1 makes labels optional, not everything. It is
  // enforced by disabling the button rather than by a toast, so that is what is
  // asserted here - submit's own titleRequired branch is unreachable from the UI.
  it('still refuses a task with no title', () => {
    labelsData.data = [{ id: 7, name: 'Bug' }];
    render(<QuickCreate />);

    const create = screen.getByRole('button', { name: 'flow.quickCreate.createTask' });
    expect(create).toBeDisabled();

    fireEvent.click(create);
    expect(createTaskMutateAsync).not.toHaveBeenCalled();
  });
});
