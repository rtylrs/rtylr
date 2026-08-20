import React from 'react';
import { render, screen, fireEvent, within } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiProject, ApiTask } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts ? `${key} ${JSON.stringify(opts)}` : key,
    i18n: { language: 'en-US' },
  }),
}));

vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => ({}) }),
  Link: ({ children }: { children?: React.ReactNode }) => <>{children}</>,
}));

const upsertDayMutate = vi.hoisted(() => vi.fn());
const upsertDayMutateAsync = vi.hoisted(() => vi.fn().mockResolvedValue({ workDay: { id: '9' } }));
const createLineMutate = vi.hoisted(() => vi.fn());
const createLineMutateAsync = vi.hoisted(() =>
  vi.fn().mockResolvedValue({ line: { id: 'line-1' } })
);
const deleteLineMutate = vi.hoisted(() => vi.fn());

const workDayState = vi.hoisted(() => ({
  value: {
    data: {
      workDay: { id: '9', workMode: 'onsite', clockedMinutes: 0, loggedMinutes: 0, gapMinutes: 0 },
      lines: [],
    } as
      | {
          workDay: Record<string, unknown>;
          lines: unknown[];
        }
      | undefined,
    isLoading: false,
  },
}));

const workDaysRangeState = vi.hoisted(() => ({
  value: { data: { workDays: [] as Array<{ date: string; loggedMinutes: number }> } },
}));

vi.mock('./useWorkDay', () => ({
  todayLocalDate: () => '2026-08-16',
  localTimezone: () => 'Asia/Beirut',
  newClientKey: () => 'client-key-1',
  useWorkDay: () => workDayState.value,
  useWorkTypes: () => ({ data: { workTypes: [] } }),
  useWorkDaysRange: () => workDaysRangeState.value,
  useUpsertWorkDay: () => ({
    mutate: upsertDayMutate,
    mutateAsync: upsertDayMutateAsync,
    isPending: false,
  }),
  useCreateWorkDayLine: () => ({
    mutate: createLineMutate,
    mutateAsync: createLineMutateAsync,
    isPending: false,
  }),
  useDeleteWorkDayLine: () => ({ mutate: deleteLineMutate, isPending: false }),
}));

const timeSettingsState = vi.hoisted(() => ({
  value: {
    weekStart: 'monday' as 'monday' | 'sunday',
    weeklyTargetMinutes: 40 * 60,
    trackAttendance: true,
  },
}));

vi.mock('../../shared/stores/timeSettingsStore', () => ({
  useTimeSettingsStore: (selector: (s: typeof timeSettingsState.value) => unknown) =>
    selector(timeSettingsState.value),
}));

const projectsState = vi.hoisted(() => ({
  value: { data: [] as ApiProject[] },
}));

vi.mock('../projects/useProjects', () => ({
  useProjects: () => projectsState.value,
}));

const tasksState = vi.hoisted(() => ({
  value: { data: [] as ApiTask[] },
}));

vi.mock('../tasks/useTasks', () => ({
  useTasks: () => tasksState.value,
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: (selector: (s: { activeWorkspaceId: string }) => unknown) =>
    selector({ activeWorkspaceId: 'ws-1' }),
}));

import MyDay from './MyDay';

function project(over: Partial<ApiProject> = {}): ApiProject {
  return {
    id: '1',
    workspaceId: 'ws-1',
    name: 'Site fitout',
    color: '#03aed2',
    taskCount: 0,
    completedCount: 0,
    memberAvatars: [],
    ...over,
  };
}

function task(over: Partial<ApiTask> = {}): ApiTask {
  return {
    id: 't1',
    projectId: '1',
    workspaceId: 'ws-1',
    title: 'Fix the gate hinge',
    status: 'todo',
    priority: 'medium',
    assignee: 'Ada',
    assigneeInitials: 'A',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Site fitout',
    ...over,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  workDayState.value = {
    data: {
      workDay: { id: '9', workMode: 'onsite', clockedMinutes: 0, loggedMinutes: 0, gapMinutes: 0 },
      lines: [],
    },
    isLoading: false,
  };
  projectsState.value = { data: [project(), project({ id: '2', name: 'Warehouse racking' })] };
  workDaysRangeState.value = { data: { workDays: [] } };
  tasksState.value = { data: [] };
  timeSettingsState.value = {
    weekStart: 'monday',
    weeklyTargetMinutes: 40 * 60,
    trackAttendance: true,
  };
});

describe('MyDay: week strip', () => {
  it('renders all seven days of the week containing the selected date', () => {
    render(<MyDay />);
    // 2026-08-16 is a Sunday; the Monday-start week runs 08-10 to 08-16.
    for (const day of ['10', '11', '12', '13', '14', '15', '16']) {
      expect(screen.getByRole('button', { name: new RegExp(`\\b${day}\\b`) })).toBeInTheDocument();
    }
  });

  it('marks a day with recorded time so it is distinguishable from an empty one', () => {
    workDaysRangeState.value = { data: { workDays: [{ date: '2026-08-11', loggedMinutes: 90 }] } };
    render(<MyDay />);
    expect(screen.getByTitle('flow.time.weekDayLogged')).toBeInTheDocument();
  });

  it('selecting a strip day changes which day is loaded', () => {
    render(<MyDay />);
    fireEvent.click(screen.getByRole('button', { name: /\b11\b/ }));
    expect(screen.getByLabelText('flow.time.date', { selector: '#work-day-date' })).toHaveAttribute(
      'data-value',
      '2026-08-11'
    );
  });
});

describe('MyDay: Add Hours renders inline', () => {
  it('renders the Add Hours form directly under What you worked on, with no dialog to open', () => {
    render(<MyDay />);
    expect(screen.getByText('flow.time.add.title')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'flow.time.logAction' })).not.toBeInTheDocument();
  });

  it('shows the workspace projects in the inline Project picker', () => {
    render(<MyDay />);
    const select = screen.getByLabelText('flow.time.project') as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.textContent)).toEqual([
      'flow.time.add.projectPlaceholder',
      'Site fitout',
      'Warehouse racking',
    ]);
  });

  it('changing the Add Hours date field moves the day strip - they are one value', () => {
    render(<MyDay />);
    fireEvent.click(screen.getByLabelText('flow.time.date', { selector: '#add-hours-date' }));
    fireEvent.click(screen.getByRole('gridcell', { name: 'Tuesday, August 11, 2026' }));
    expect(screen.getByLabelText('flow.time.date', { selector: '#work-day-date' })).toHaveAttribute(
      'data-value',
      '2026-08-11'
    );
  });
});

describe('MyDay: trackAttendance gating', () => {
  it('shows the attendance inputs and the three-figure strip by default', () => {
    render(<MyDay />);
    expect(screen.getByText('flow.time.theDay')).toBeInTheDocument();
    expect(screen.getByLabelText('flow.time.workMode')).toBeInTheDocument();
    expect(screen.getByText('flow.time.clocked')).toBeInTheDocument();
    expect(screen.getByText('flow.time.logged')).toBeInTheDocument();
    expect(screen.getByText('flow.time.unaccounted')).toBeInTheDocument();
  });

  it('keeps Location visible but hides clock inputs and collapses to a single Logged figure when off', () => {
    timeSettingsState.value = { ...timeSettingsState.value, trackAttendance: false };
    render(<MyDay />);
    // Location is a separate concept from shift attendance - Add Hours' own
    // per-entry Location silently defaults to this value, so "The day" and
    // its Where-you-worked toggle must stay visible even with tracking off.
    expect(screen.getByText('flow.time.theDay')).toBeInTheDocument();
    expect(screen.getByLabelText('flow.time.workMode')).toBeInTheDocument();
    expect(screen.queryByLabelText('flow.time.startTime')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('flow.time.endTime')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('flow.time.breakMinutes')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.time.clocked')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.time.unaccounted')).not.toBeInTheDocument();
    expect(screen.getByText('flow.time.logged')).toBeInTheDocument();
  });
});

describe('MyDay: entry list labels', () => {
  it("shows the task's identifier and title for a line logged against a task", () => {
    tasksState.value = { data: [task({ identifier: 'T-1' })] };
    workDayState.value = {
      data: {
        workDay: {
          id: '9',
          workMode: 'onsite',
          clockedMinutes: 0,
          loggedMinutes: 90,
          gapMinutes: 0,
        },
        lines: [
          { id: 'line-1', durationMinutes: 90, description: '-', projectId: '1', taskId: 't1' },
        ],
      },
      isLoading: false,
    };
    render(<MyDay />);
    expect(screen.getByText('T-1 Fix the gate hinge')).toBeInTheDocument();
  });

  it('falls back to the project name for a line logged with no task', () => {
    workDayState.value = {
      data: {
        workDay: {
          id: '9',
          workMode: 'onsite',
          clockedMinutes: 0,
          loggedMinutes: 90,
          gapMinutes: 0,
        },
        lines: [{ id: 'line-1', durationMinutes: 90, description: '-', projectId: '2' }],
      },
      isLoading: false,
    };
    render(<MyDay />);
    expect(within(screen.getByRole('list')).getByText('Warehouse racking')).toBeInTheDocument();
  });
});
