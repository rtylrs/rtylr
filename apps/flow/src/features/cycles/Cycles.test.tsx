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

const cyclesState = vi.hoisted(() => ({
  value: { data: [] as ApiCycle[], isLoading: false },
}));

// The screen asks which project's cycles to show, so it reads the project list.
vi.mock('../projects/useProjects', () => ({
  useProjects: () => ({ data: [{ id: 'p-1', name: 'Website Redesign' }], isLoading: false }),
}));

vi.mock('./useCycles', () => ({
  useCycles: () => cyclesState.value,
  useCreateCycle: () => ({ mutate: mutateSpy, isPending: false }),
}));

import Cycles from './Cycles';

function cycle(over: Partial<ApiCycle> = {}): ApiCycle {
  return {
    id: '1',
    workspaceId: 'ws-1',
    name: 'Sprint 1',
    startDate: '2026-08-01',
    endDate: '2026-08-14',
    status: 'active',
    scope: 10,
    completed: 4,
    ...over,
  };
}

describe('Cycles', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    cyclesState.value = { data: [], isLoading: false };
  });

  it('shows a real empty state with a create CTA when the workspace has zero cycles', () => {
    cyclesState.value = { data: [], isLoading: false };
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
});

describe('Cycles create form', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-08-20T12:00:00.000Z'));
    vi.clearAllMocks();
    cyclesState.value = { data: [], isLoading: false };
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
