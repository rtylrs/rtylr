import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { SavedView } from '../../shared/stores/flowStore';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

const viewsState = vi.hoisted(() => ({ data: [] as SavedView[] }));
const createViewMutate = vi.hoisted(() => vi.fn());
const updateViewMutate = vi.hoisted(() => vi.fn());
const filteredTasksState = vi.hoisted(() => ({
  value: [] as Array<{ id: string; title: string }>,
}));

vi.mock('./useViews', () => ({
  useViews: () => viewsState,
  useCreateView: () => ({ mutate: createViewMutate, isPending: false }),
  useUpdateView: () => ({ mutate: updateViewMutate, isPending: false }),
  useDeleteView: () => ({ mutate: vi.fn(), isPending: false }),
  useFilteredTasks: () => filteredTasksState.value,
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ tasks: [] }),
}));

vi.mock('../../shared/components/TaskRow', () => ({
  default: ({ task }: { task: { title: string } }) => <div>{task.title}</div>,
}));

vi.mock('../../shared/components/ConfirmDialog', () => ({
  default: () => null,
}));

import Views from './Views';

beforeEach(() => {
  vi.clearAllMocks();
  viewsState.data = [];
  filteredTasksState.value = [];
});

describe('Views: no saved views yet', () => {
  it('shows the empty title with a CTA that creates the first view', () => {
    render(<Views />);

    expect(screen.getByText('flow.views.emptyList')).toBeInTheDocument();
    const createButtons = screen.getAllByRole('button', { name: 'flow.views.addTooltip' });
    expect(createButtons.length).toBeGreaterThan(0);

    fireEvent.click(createButtons[createButtons.length - 1]);
    expect(createViewMutate).toHaveBeenCalled();
  });
});

describe('Views: a view with no matching tasks', () => {
  it('shows the empty-tasks title with a CTA to My Tasks', () => {
    viewsState.data = [{ id: 'v1', name: 'Urgent', icon: 'flame', filters: {} } as SavedView];
    filteredTasksState.value = [];

    render(<Views />);

    expect(screen.getByText('flow.views.emptyTasksForView')).toBeInTheDocument();
    const cta = screen.getByRole('button', { name: 'flow.shell.myTasks' });
    fireEvent.click(cta);
    expect(navigateMock).toHaveBeenCalledWith({ to: '/my-issues' });
  });
});

describe('Views: per-view visibility (S2)', () => {
  const view = (over: Partial<SavedView> = {}): SavedView =>
    ({ id: '1', name: 'My view', icon: 'eye', filters: {}, ...over }) as SavedView;

  // An absent visibility must read as SHARED. That is the column default and how
  // every view behaved before the field existed, so treating it as private would
  // put a lock on views the whole business can actually see.
  it('treats a view with no visibility as shared', () => {
    viewsState.data = [view()];

    render(<Views />);

    expect(screen.getByLabelText('flow.views.makePrivateLabel')).toBeInTheDocument();
    expect(screen.queryByLabelText('flow.views.makeSharedLabel')).not.toBeInTheDocument();
  });

  it('offers to re-share a view that is already private', () => {
    viewsState.data = [view({ visibility: 'private' })];

    render(<Views />);

    expect(screen.getByLabelText('flow.views.makeSharedLabel')).toBeInTheDocument();
  });

  it('sends private when toggling a shared view', () => {
    viewsState.data = [view()];

    render(<Views />);
    fireEvent.click(screen.getByLabelText('flow.views.makePrivateLabel'));

    expect(updateViewMutate).toHaveBeenCalledWith(
      { id: '1', data: { visibility: 'private' } },
      expect.anything()
    );
  });

  it('sends shared when toggling a private view back', () => {
    viewsState.data = [view({ visibility: 'private' })];

    render(<Views />);
    fireEvent.click(screen.getByLabelText('flow.views.makeSharedLabel'));

    expect(updateViewMutate).toHaveBeenCalledWith(
      { id: '1', data: { visibility: 'shared' } },
      expect.anything()
    );
  });
});
