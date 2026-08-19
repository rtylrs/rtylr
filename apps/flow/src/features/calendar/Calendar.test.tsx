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
  useFlowStore: () => ({ activeWorkspaceId: 'ws-1', projects: [{ id: 'p-9', name: 'Web' }] }),
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
