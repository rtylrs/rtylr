import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { ApiCycle } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      if (key === 'flow.cycles.taskCount' && opts) return `${opts.done} of ${opts.total} tasks`;
      if (key === 'flow.cycles.dateRange' && opts) return `${opts.start} to ${opts.end}`;
      return key;
    },
    i18n: { language: 'en-US' },
  }),
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({ activeWorkspaceId: 'ws-1' }),
}));

const mutateSpy = vi.fn();

interface CyclesQueryState {
  data: ApiCycle[];
  isLoading: boolean;
  isPending?: boolean;
  isError?: boolean;
  isPaused?: boolean;
  failureCount?: number;
  refetch?: () => unknown;
}

const cyclesState = vi.hoisted(() => ({
  value: { data: [] as ApiCycle[], isLoading: false, isPending: false } as CyclesQueryState,
}));

// The screen asks which project's cycles to show, so it reads the project list.
const projectsState = vi.hoisted(() => ({
  data: [{ id: 'p-1', name: 'Website Redesign' }] as Array<{ id: string; name: string }>,
}));
vi.mock('../projects/useProjects', () => ({
  useProjects: () => ({ data: projectsState.data, isLoading: false }),
}));

vi.mock('./useCycles', () => ({
  useCycles: () => cyclesState.value,
  useCreateCycle: () => ({ mutate: mutateSpy, isPending: false }),
}));

import Cycles from './Cycles';

function cycle(over: Partial<ApiCycle> = {}): ApiCycle {
  return {
    id: '1',
    name: 'Sprint 1',
    startDate: '2026-08-01',
    endDate: '2026-08-14',
    status: 'active',
    scope: 10,
    completed: 4,
    ...over,
  };
}

/**
 * Every describe here shares one mocked project list, and the last one empties
 * it to check the disabled primary. Nothing put it back, so under
 * `--sequence.shuffle` the create-form tests could render with no projects,
 * where the form is gated shut, and failed looking for a date picker that was
 * never on screen (#1194).
 *
 * Top-level, so a describe added later inherits the reset rather than having
 * to remember it.
 */
beforeEach(() => {
  projectsState.data = [{ id: 'p-1', name: 'Website Redesign' }];
});

describe('Cycles', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cyclesState.value = { data: [], isLoading: false, isPending: false };
  });

  it('shows a real empty state with a create CTA when the workspace has zero cycles', () => {
    cyclesState.value = { data: [], isLoading: false, isPending: false };
    render(<Cycles />);

    expect(screen.getByText('flow.cycles.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.cycles.emptyDescription')).toBeInTheDocument();
    expect(screen.getByText('flow.cycles.createFirst')).toBeInTheDocument();
  });

  it('renders the active, upcoming and completed cycles instead of the empty state when cycles exist', () => {
    cyclesState.value = {
      data: [
        cycle({ id: '1', name: 'Active Sprint', status: 'active' }),
        cycle({ id: '2', name: 'Next Sprint', status: 'upcoming' }),
        cycle({ id: '3', name: 'Past Sprint', status: 'completed' }),
      ],
      isLoading: false,
    };
    render(<Cycles />);

    expect(screen.queryByText('flow.cycles.emptyTitle')).not.toBeInTheDocument();
    // Each name renders twice: once in the section row, once inside CycleBadge.
    expect(screen.getAllByText('Active Sprint').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Next Sprint').length).toBeGreaterThan(0);
    expect(screen.getAllByText('Past Sprint').length).toBeGreaterThan(0);
  });

  it('keeps loading while a project query is paused before its first attempt', () => {
    cyclesState.value = {
      data: [],
      isLoading: false,
      isPending: true,
      isPaused: true,
      failureCount: 0,
    };
    render(<Cycles />);

    expect(screen.getByText('flow.cycles.loading')).toBeInTheDocument();
    expect(screen.queryByText('flow.cycles.emptyTitle')).not.toBeInTheDocument();
  });

  it('renders every active cycle when the API returns more than one', () => {
    cyclesState.value = {
      data: [
        cycle({ id: '1', name: 'First Active', status: 'active' }),
        cycle({ id: '2', name: 'Second Active', status: 'active' }),
      ],
      isLoading: false,
    };
    render(<Cycles />);

    expect(screen.getByText('First Active')).toBeInTheDocument();
    expect(screen.getByText('Second Active')).toBeInTheDocument();
  });
});

describe('Cycles create form', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-08-20T12:00:00.000Z'));
    vi.clearAllMocks();
    cyclesState.value = { data: [], isLoading: false, isPending: false };
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('gives the start and end date pickers accessible names that open a calendar', () => {
    render(<Cycles />);

    fireEvent.click(screen.getByText('flow.cycles.newCycle'));

    fireEvent.click(screen.getByLabelText('flow.cycles.startDate'));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('gridcell', { name: 'Thursday, August 20, 2026' }));

    fireEvent.click(screen.getByLabelText('flow.cycles.endDate'));
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('submits the exact start and end dates picked by the user', () => {
    render(<Cycles />);

    fireEvent.click(screen.getByText('flow.cycles.newCycle'));
    fireEvent.change(screen.getByPlaceholderText('flow.cycles.namePlaceholder'), {
      target: { value: 'Sprint 1' },
    });

    fireEvent.click(screen.getByLabelText('flow.cycles.startDate'));
    fireEvent.click(screen.getByRole('gridcell', { name: 'Thursday, August 20, 2026' }));

    fireEvent.click(screen.getByLabelText('flow.cycles.endDate'));
    fireEvent.click(screen.getByRole('gridcell', { name: 'Friday, August 21, 2026' }));

    fireEvent.click(screen.getByText('flow.cycles.create'));

    expect(mutateSpy).toHaveBeenCalledTimes(1);
    expect(mutateSpy.mock.calls[0][0]).toMatchObject({
      name: 'Sprint 1',
      startDate: '2026-08-20',
      endDate: '2026-08-21',
    });
  });
});

describe('Cycles: the disabled primary explains itself (#1090)', () => {
  beforeEach(() => {
    projectsState.data = [{ id: 'p-1', name: 'Website Redesign' }];
  });

  it('says why New cycle is disabled when there is no project', () => {
    // A cycle belongs to a project, so with none there is nothing to create
    // one against. The button was simply dim.
    projectsState.data = [];
    render(<Cycles />);

    expect(screen.getByText('flow.cycles.newCycleGate')).toBeInTheDocument();
    expect(screen.getByText('flow.cycles.newCycle').closest('button')).toBeDisabled();
  });

  it('points the button at the reason for a screen reader', () => {
    projectsState.data = [];
    render(<Cycles />);

    expect(screen.getByText('flow.cycles.newCycle').closest('button')).toHaveAttribute(
      'aria-describedby',
      'cycles-project-gate'
    );
  });

  it('explains nothing when there is a project to create against', () => {
    render(<Cycles />);

    expect(screen.queryByText('flow.cycles.newCycleGate')).not.toBeInTheDocument();
    expect(screen.getByText('flow.cycles.newCycle').closest('button')).not.toBeDisabled();
  });

  it('shows the project label rather than hiding it from sighted users', () => {
    // It was sr-only, on a selector that truncates a 60-character project
    // name, which is exactly where the reader needs to know what it is.
    render(<Cycles />);
    const label = screen.getByText('flow.cycles.project');
    expect(label.className).not.toContain('sr-only');
  });
});
