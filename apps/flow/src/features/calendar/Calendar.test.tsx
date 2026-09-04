import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts ? `${key} ${JSON.stringify(opts)}` : key,
    i18n: { language: 'en-US' },
  }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

const tasksData = vi.hoisted(() => ({ data: [] as unknown[] }));
const createTaskMutate = vi.hoisted(() => vi.fn());
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => tasksData,
  useCreateTask: () => ({ mutate: createTaskMutate, isPending: false }),
}));

const labelsData = vi.hoisted(() => ({ data: [] as { id: number; name: string }[] }));
vi.mock('../tasks/useLabels', () => ({
  useLabels: () => labelsData,
}));

vi.mock('../../shared/stores/flowStore', () => ({
  // activeWorkspaceId is '' in production, since the workspace layer was
  // removed. The old submit guard required it, so a mock carrying 'ws-1' would
  // hide the very defect these tests cover.
  useFlowStore: () => ({ activeWorkspaceId: '', projects: storeProjects.value }),
}));

const storeProjects = vi.hoisted(() => ({
  value: [{ id: 'p-9', name: 'Web' }] as { id: string; name: string }[],
}));

const mockNavigate = vi.fn();
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => mockNavigate,
}));

import Calendar from './Calendar';

beforeEach(() => {
  vi.clearAllMocks();
  tasksData.data = [];
  labelsData.data = [];
  storeProjects.value = [{ id: 'p-9', name: 'Web' }];
});

describe('Calendar per-day add-task control', () => {
  it('renders the add-task affordance as a real, accessibly-named button', () => {
    render(<Calendar />);

    const addButtons = screen.getAllByRole('button', {
      name: /flow\.calendar\.addTaskLabel/,
    });

    expect(addButtons.length).toBeGreaterThan(0);
    expect(addButtons[0].tagName).toBe('BUTTON');
    // The accessible name carries the concrete date, not just the raw key.
    expect(addButtons[0]).toHaveAccessibleName(/"date":"/);
  });

  it('is reachable via Tab and opens the create-task modal on click', () => {
    render(<Calendar />);

    const [addButton] = screen.getAllByRole('button', {
      name: /flow\.calendar\.addTaskLabel/,
    });

    // Native <button> elements are focusable without extra tabIndex wiring.
    addButton.focus();
    expect(addButton).toHaveFocus();

    fireEvent.click(addButton);

    expect(screen.getByText('flow.calendar.newTaskFor')).toBeInTheDocument();
  });

  it('opens the create-task modal when activated with the Enter key', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const [addButton] = screen.getAllByRole('button', {
      name: /flow\.calendar\.addTaskLabel/,
    });

    addButton.focus();
    await user.keyboard('{Enter}');

    expect(screen.getByText('flow.calendar.newTaskFor')).toBeInTheDocument();
  });

  it('opens the create-task modal when activated with the Space key', async () => {
    const user = userEvent.setup();
    render(<Calendar />);

    const [addButton] = screen.getAllByRole('button', {
      name: /flow\.calendar\.addTaskLabel/,
    });

    addButton.focus();
    await user.keyboard(' ');

    expect(screen.getByText('flow.calendar.newTaskFor')).toBeInTheDocument();
  });

  it('does not double-open the modal via the surrounding day cell click handler', () => {
    render(<Calendar />);

    const [addButton] = screen.getAllByRole('button', {
      name: /flow\.calendar\.addTaskLabel/,
    });

    fireEvent.click(addButton);

    // Only one modal instance should be present.
    expect(screen.getAllByText('flow.calendar.newTaskFor')).toHaveLength(1);
  });
});

describe('Calendar day-add submission (#1057)', () => {
  const openModalAndType = (title: string): void => {
    const [addButton] = screen.getAllByRole('button', {
      name: /flow\.calendar\.addTaskLabel/,
    });
    fireEvent.click(addButton);
    fireEvent.change(screen.getByPlaceholderText(/flow\.calendar\.titlePlaceholder/), {
      target: { value: title },
    });
  };

  const submit = (): void => {
    fireEvent.click(screen.getByRole('button', { name: 'flow.calendar.create' }));
  };

  it('creates the task with the day as its due date', () => {
    render(<Calendar />);

    openModalAndType('Write the brief');
    submit();

    expect(createTaskMutate).toHaveBeenCalledTimes(1);
    const payload = createTaskMutate.mock.calls[0][0] as Record<string, unknown>;
    expect(payload).toMatchObject({ projectId: 'p-9', title: 'Write the brief' });
    expect(payload.dueDate).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });

  it('creates a task when the business has no labels at all', () => {
    labelsData.data = [];
    render(<Calendar />);

    openModalAndType('No labels here');
    submit();

    // The guard used to require a label, so this refused outright.
    expect(createTaskMutate).toHaveBeenCalledTimes(1);
    expect(createTaskMutate.mock.calls[0][0]).not.toHaveProperty('labels');
  });

  it('sends the label when one is available and pre-selected', () => {
    labelsData.data = [{ id: 994343, name: 'qa-temp' }];
    render(<Calendar />);

    openModalAndType('With a label');
    submit();

    expect(createTaskMutate).toHaveBeenCalledTimes(1);
    expect(createTaskMutate.mock.calls[0][0]).toMatchObject({ labels: ['994343'] });
  });

  it('still refuses when the business has no project to create into', async () => {
    const { toast } = await import('sonner');
    storeProjects.value = [];
    render(<Calendar />);

    openModalAndType('Nowhere to put this');
    submit();

    expect(createTaskMutate).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith('toasts.task.selectProjectFirst');
  });
});
