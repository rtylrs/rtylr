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
const filteredTasksState = vi.hoisted(() => ({
  value: [] as Array<{ id: string; title: string }>,
}));

vi.mock('./useViews', () => ({
  useViews: () => viewsState,
  useCreateView: () => ({ mutate: createViewMutate, isPending: false }),
  useUpdateView: () => ({ mutate: vi.fn(), isPending: false }),
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
