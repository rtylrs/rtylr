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
    isError: false,
    isPaused: false,
    failureCount: 0,
    refetch: vi.fn(),
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
    dailyTargetMinutes: {
      sun: 0,
      mon: 8 * 60,
      tue: 8 * 60,
      wed: 8 * 60,
      thu: 8 * 60,
      fri: 8 * 60,
      sat: 4 * 60,
    } as Record<string, number>,
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
    isError: false,
    isPaused: false,
    failureCount: 0,
    refetch: vi.fn(),
  };
  projectsState.value = { data: [project(), project({ id: '2', name: 'Warehouse racking' })] };
  workDaysRangeState.value = { data: { workDays: [] } };
  tasksState.value = { data: [] };
  timeSettingsState.value = {
    weekStart: 'monday',
    dailyTargetMinutes: {
      sun: 0,
      mon: 8 * 60,
      tue: 8 * 60,
      wed: 8 * 60,
      thu: 8 * 60,
      fri: 8 * 60,
      sat: 4 * 60,
    },
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
    // This asserted a dot carrying only a `title`, which is aria-hidden and
    // announced as nothing. The cell now shows the total itself, which is a
    // stronger affordance and readable by anything.
    workDaysRangeState.value = { data: { workDays: [{ date: '2026-08-11', loggedMinutes: 90 }] } };
    render(<MyDay />);
    expect(screen.getByText('1:30')).toBeInTheDocument();
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

describe('MyDay: no work lines yet', () => {
  it('shows the empty title with an Add hours CTA that scrolls to the inline form', () => {
    render(<MyDay />);

    expect(screen.getByText('flow.time.noLinesYet')).toBeInTheDocument();
    const cta = screen.getByRole('button', { name: 'flow.time.noLinesYetCta' });
    expect(cta).toBeEnabled();

    fireEvent.click(cta);
    expect(Element.prototype.scrollIntoView).toHaveBeenCalled();
  });
});

describe('MyDay: day-level attendance and summary controls removed', () => {
  it('shows no Where-you-worked picker, clock inputs, or standalone Logged figure', () => {
    render(<MyDay />);
    expect(screen.queryByText('flow.time.theDay')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('flow.time.workMode')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('flow.time.startTime')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('flow.time.endTime')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('flow.time.breakMinutes')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.time.clocked')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.time.unaccounted')).not.toBeInTheDocument();
    // Add Hours' own "Today X of Y" pill already shows this, with the
    // added context of the day's target - a second, bare copy here was
    // dead weight (and a full-width empty bar once the other two figures
    // it used to sit beside were removed).
    expect(screen.queryByText('flow.time.logged')).not.toBeInTheDocument();
  });

  it('stays that way regardless of the trackAttendance setting', () => {
    timeSettingsState.value = { ...timeSettingsState.value, trackAttendance: false };
    render(<MyDay />);
    expect(screen.queryByText('flow.time.theDay')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.time.logged')).not.toBeInTheDocument();
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

describe('MyDay: a failed load is not an endless load', () => {
  it('shows the failure state instead of loading forever', () => {
    // The reported case: /time sat on "Loading the day..." indefinitely,
    // because a failed query is not a loading one and nothing rendered the
    // difference.
    workDayState.value = { ...workDayState.value, isError: true };
    render(<MyDay />);
    expect(screen.getByText('flow.loadFailed.title')).toBeInTheDocument();
    expect(screen.queryByText('flow.time.loadingDay')).not.toBeInTheDocument();
  });

  it('retries the query from the failure state', () => {
    const refetch = vi.fn();
    workDayState.value = { ...workDayState.value, isError: true, refetch };
    render(<MyDay />);
    fireEvent.click(screen.getByText('flow.loadFailed.retry'));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it('still shows the loading copy while genuinely loading', () => {
    workDayState.value = { ...workDayState.value, isLoading: true };
    render(<MyDay />);
    expect(screen.getByText('flow.time.loadingDay')).toBeInTheDocument();
    expect(screen.queryByText('flow.loadFailed.title')).not.toBeInTheDocument();
  });
});

describe('MyDay day total and week strip (#1093)', () => {
  it('shows the day total prominently, not only in the card pill', () => {
    // S-19 asks for 20px where the eye lands. It was 12px inside a pill at
    // the top right of the Add-hours card.
    workDayState.value = {
      ...workDayState.value,
      data: {
        workDay: {
          id: '9',
          workMode: 'onsite',
          clockedMinutes: 0,
          loggedMinutes: 185,
          gapMinutes: 0,
        },
        lines: [],
      },
    };
    const { container } = render(<MyDay />);

    const header = container.querySelector('header');
    expect(header?.textContent).toContain('3:05');
    // The prominent number is on the type scale's largest step.
    const prominent = [...(header?.querySelectorAll('.text-flow-screen') ?? [])].map(
      (n) => n.textContent
    );
    expect(prominent.some((text) => text?.includes('3:05'))).toBe(true);
  });

  it('shows each day of the week its own logged total', () => {
    // The strip already fetched loggedMinutes per date and used it only to
    // decide whether to draw a dot, so every cell carried just the date.
    workDaysRangeState.value = {
      data: { workDays: [{ date: '2026-08-16', loggedMinutes: 125 }] },
    };
    render(<MyDay />);

    expect(screen.getByText('2:05')).toBeInTheDocument();
  });

  it('marks a day with nothing logged rather than leaving it blank', () => {
    workDaysRangeState.value = { data: { workDays: [] } };
    const { container } = render(<MyDay />);

    const strip = container.querySelector('[role="group"]');
    expect(strip?.textContent).toContain('–');
  });

  it('omits the target when none is configured', () => {
    workDayState.value = {
      ...workDayState.value,
      data: {
        workDay: {
          id: '9',
          workMode: 'onsite',
          clockedMinutes: 0,
          loggedMinutes: 60,
          gapMinutes: 0,
        },
        lines: [],
      },
    };
    const { container } = render(<MyDay />);
    expect(container.querySelector('header')?.textContent).not.toContain('flow.time.of');
  });
});
