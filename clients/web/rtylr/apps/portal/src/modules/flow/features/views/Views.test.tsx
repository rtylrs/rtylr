import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { createInstance } from 'i18next';
import { I18nextProvider } from 'react-i18next';
import flow from '../../locales/en/flow.json';
import flowCommon from '../../locales/en/flowCommon.json';
import type { ApiTask } from '../../shared/services/flowApi';
import type { SavedView } from '../../shared/stores/flowStore';

const i18n = createInstance();
void i18n.init({
  lng: 'en',
  resources: { en: { translation: { flow, flowCommon } } },
  interpolation: { escapeValue: false },
});
const renderViews = () =>
  render(
    <I18nextProvider i18n={i18n}>
      <Views />
    </I18nextProvider>
  );

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
const tasksQuery = vi.hoisted(() => ({
  data: [] as ApiTask[],
  isLoading: false,
  isError: false,
  refetch: vi.fn(),
}));
vi.mock('../tasks/useTasks', () => ({ useTasks: () => tasksQuery }));

vi.mock('./useViews', async (importOriginal) => ({
  ...(await importOriginal<typeof import('./useViews')>()),
  useViews: () => viewsState,
  useCreateView: () => ({ mutate: createViewMutate, isPending: false }),
  useUpdateView: () => ({ mutate: updateViewMutate, isPending: false }),
  useDeleteView: () => ({ mutate: vi.fn(), isPending: false }),
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
  tasksQuery.data = [];
  tasksQuery.isLoading = false;
  tasksQuery.isError = false;
});

describe('Views: no saved views yet', () => {
  it('shows the empty title with a CTA that creates the first view', () => {
    renderViews();

    expect(screen.getByText('No saved views yet. Click + to create one.')).toBeInTheDocument();
    const createButtons = screen.getAllByRole('button', { name: 'New view' });
    expect(createButtons.length).toBeGreaterThan(0);

    fireEvent.click(createButtons[createButtons.length - 1]);
    expect(createViewMutate).toHaveBeenCalled();
  });
});

describe('Views: a view with no matching tasks', () => {
  it('shows the empty-tasks title with a CTA to My Tasks', () => {
    viewsState.data = [{ id: 'v1', name: 'Urgent', icon: 'flame', filters: {} } as SavedView];
    tasksQuery.data = [];
    tasksQuery.isLoading = false;
    tasksQuery.isError = false;

    renderViews();

    expect(screen.getByText("No tasks match this view's filters.")).toBeInTheDocument();
    const cta = screen.getByRole('button', { name: 'My tasks' });
    fireEvent.click(cta);
    expect(navigateMock).toHaveBeenCalledWith({ to: '/flow/my-issues' });
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

    renderViews();

    expect(screen.getByLabelText('Make view private')).toBeInTheDocument();
    expect(screen.queryByLabelText('Share view')).not.toBeInTheDocument();
  });

  it('offers to re-share a view that is already private', () => {
    viewsState.data = [view({ visibility: 'private' })];

    renderViews();

    expect(screen.getByLabelText('Share view')).toBeInTheDocument();
  });

  it('sends private when toggling a shared view', () => {
    viewsState.data = [view()];

    renderViews();
    fireEvent.click(screen.getByLabelText('Make view private'));

    expect(updateViewMutate).toHaveBeenCalledWith(
      { id: '1', data: { visibility: 'private' } },
      expect.anything()
    );
  });

  it('sends shared when toggling a private view back', () => {
    viewsState.data = [view({ visibility: 'private' })];

    renderViews();
    fireEvent.click(screen.getByLabelText('Share view'));

    expect(updateViewMutate).toHaveBeenCalledWith(
      { id: '1', data: { visibility: 'shared' } },
      expect.anything()
    );
  });
});

describe('Views task loading', () => {
  const task = (id: string, status: ApiTask['status'], archived = false): ApiTask => ({
    id,
    projectId: '2',
    title: `Remote task ${id}`,
    status,
    priority: 'high',
    assignee: '',
    assigneeInitials: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Project',
    archived,
  });

  it('renders remote tasks, excludes archived tasks and uses updated saved filters', () => {
    viewsState.data = [{ id: '1', name: 'Active', icon: 'eye', filters: { status: ['todo'] } }];
    tasksQuery.data = [task('1', 'todo'), task('2', 'done'), task('3', 'todo', true)];
    const rendered = renderViews();
    fireEvent.click(screen.getByRole('button', { name: 'Active' }));
    expect(screen.getByText('Remote task 1')).toBeInTheDocument();
    expect(screen.queryByText('Remote task 2')).not.toBeInTheDocument();
    expect(screen.queryByText('Remote task 3')).not.toBeInTheDocument();
    viewsState.data = [{ ...viewsState.data[0], filters: { status: ['done'] } }];
    rendered.rerender(
      <I18nextProvider i18n={i18n}>
        <Views />
      </I18nextProvider>
    );
    expect(screen.getByText('Remote task 2')).toBeInTheDocument();
    expect(screen.queryByText('Remote task 1')).not.toBeInTheDocument();
  });

  it('reports a task load failure and retries without claiming there are no matches', () => {
    viewsState.data = [{ id: '1', name: 'Active', icon: 'eye', filters: {} }];
    tasksQuery.isError = true;
    renderViews();
    expect(screen.getByText("Couldn't load this")).toBeInTheDocument();
    expect(screen.queryByText("No tasks match this view's filters.")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
    expect(tasksQuery.refetch).toHaveBeenCalledOnce();
  });
});
