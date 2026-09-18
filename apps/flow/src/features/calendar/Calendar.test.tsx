// @vitest-environment jsdom
//
// jsdom rather than the suite default. happy-dom silently discards any
// color-mix() assigned through the style property, which is the path React uses,
// so a tint is simply absent from the DOM there (rtylr#1249). The tint
// assertions below passed under happy-dom only because the same element also
// carried the raw colour as a plain hex, which was the defect fixed in
// rtylr#1246.
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  setViewportWidth,
  resetViewport,
  MOBILE_WIDTH,
  DESKTOP_WIDTH,
} from '../../../../../test/viewport';

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
  // PriorityIcon reads this, so it is only needed once a test renders a task.
  PRIORITY_COLORS: {
    urgent: '#ef4444',
    high: '#f97316',
    medium: 'var(--flow-info-fg)',
    low: '#6b7280',
  },
}));

const storeProjects = vi.hoisted(() => ({
  value: [{ id: 'p-9', name: 'Web' }] as { id: string; name: string; color?: string }[],
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

describe('Calendar pills carry their project (#1087)', () => {
  const today = new Date();
  const iso = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

  afterEach(() => {
    tasksData.data = [];
    storeProjects.value = [{ id: 'p-9', name: 'Web' }];
  });

  it('gives two projects two different tints', () => {
    // This calendar mixes every project together, so one card colour left no
    // way to tell whose work a day was full of.
    storeProjects.value = [
      { id: 'p-1', name: 'Alpha', color: '#112233' },
      { id: 'p-2', name: 'Beta', color: '#445566' },
    ];
    tasksData.data = [
      { id: 't1', title: 'From Alpha', projectId: 'p-1', dueDate: iso, priority: 'medium' },
      { id: 't2', title: 'From Beta', projectId: 'p-2', dueDate: iso, priority: 'medium' },
    ];
    render(<Calendar />);

    const alpha = screen.getByText('From Alpha').closest('button');
    const beta = screen.getByText('From Beta').closest('button');
    // The tone, which is what carries the information.
    expect(alpha?.style.getPropertyValue('--tone')).toBe('#112233');
    expect(beta?.style.getPropertyValue('--tone')).toBe('#445566');
    // And that the tone is actually spent on a tint. This is the half that was
    // never verified: under happy-dom the background was absent, and the old
    // assertion passed on a sibling `color: #112233` that was itself the bug.
    expect(alpha?.style.background).toBe('color-mix(in srgb, var(--tone) 13%, transparent)');
    expect(alpha?.style.color).toBe('color-mix(in srgb, var(--tone) 50%, var(--flow-text))');
    expect(alpha?.style.color).not.toContain('112233');
  });

  it('falls back to the card surface when the project is not in the store', () => {
    // A task can arrive before its project does, and an undefined colour must
    // not paint the pill with the string "undefined".
    storeProjects.value = [];
    tasksData.data = [
      { id: 't3', title: 'Orphan', projectId: 'p-404', dueDate: iso, priority: 'medium' },
    ];
    render(<Calendar />);

    const style = screen.getByText('Orphan').closest('button')?.getAttribute('style') ?? '';
    expect(style).toContain('flow-card');
    expect(style).not.toContain('undefined');
  });

  it('lets a right-to-left title decide its own truncation end', () => {
    tasksData.data = [
      { id: 't4', title: 'مهمة طويلة جدا', projectId: 'p-9', dueDate: iso, priority: 'medium' },
    ];
    render(<Calendar />);
    expect(screen.getByText('مهمة طويلة جدا')).toHaveAttribute('dir', 'auto');
  });
});

describe('Calendar at the 390px floor (#1215)', () => {
  const today = new Date();
  const iso = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

  /**
   * The point of rtylr#1215, and the first test in this repo able to state it.
   *
   * `matchMedia` was mocked to report `matches: false` for every query, so any
   * responsive branch rendered its no-match side and no assertion could tell
   * that apart from the viewport genuinely being narrow. The shim now evaluates
   * width features; `test/viewport.ts` has the measurements.
   *
   * A width has to be set before rendering: neither test DOM fires the media
   * query `change` event, so a component only ever sees the width it mounted at.
   */
  afterEach(() => {
    resetViewport();
    tasksData.data = [];
    storeProjects.value = [{ id: 'p-9', name: 'Web' }];
  });

  it('shows the month as a list, not a seven-column grid', () => {
    setViewportWidth(MOBILE_WIDTH);
    tasksData.data = [
      { id: 't1', title: 'From Alpha', projectId: 'p-1', dueDate: iso, priority: 'medium' },
    ];
    render(<Calendar />);

    // The agenda's day heading, which the grid has no equivalent of.
    expect(screen.getAllByRole('heading', { level: 3 }).length).toBeGreaterThan(0);
    // And the grid's weekday header row is gone rather than merely hidden.
    expect(screen.queryByText('Sun')).toBeNull();
  });

  it('keeps the grid at desktop width', () => {
    setViewportWidth(DESKTOP_WIDTH);
    tasksData.data = [
      { id: 't1', title: 'From Alpha', projectId: 'p-1', dueDate: iso, priority: 'medium' },
    ];
    render(<Calendar />);

    expect(screen.getByText('Sun')).toBeInTheDocument();
    expect(screen.queryAllByRole('heading', { level: 3 })).toHaveLength(0);
  });

  it('opens the task over the list rather than replacing it', () => {
    setViewportWidth(MOBILE_WIDTH);
    tasksData.data = [
      { id: 't1', title: 'From Alpha', projectId: 'p-1', dueDate: iso, priority: 'medium' },
    ];
    render(<Calendar />);

    screen.getByText('From Alpha').closest('button')!.click();

    // `?openTask`, not `/task/t1`. The calendar stays mounted underneath and
    // the drawer opens over it, so coming back does not re-fetch and re-scroll
    // the month (rtylr#1079). `to: '.'` keeps the current route.
    expect(mockNavigate).toHaveBeenCalledTimes(1);
    const call = mockNavigate.mock.calls[0][0] as {
      to: string;
      search: (previous: Record<string, unknown>) => Record<string, unknown>;
    };
    expect(call.to).toBe('.');
    // The updater is a function, so that opening a task cannot discard the
    // filters the screen underneath already has in its query string.
    expect(call.search({ view: 'month' })).toEqual({ view: 'month', openTask: 't1' });
  });
});
