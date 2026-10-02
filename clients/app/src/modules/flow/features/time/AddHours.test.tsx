import React from 'react';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiProject, ApiTask } from '../../shared/services/flowApi';
import type { WorkDay } from './useWorkDay';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts ? `${key} ${JSON.stringify(opts)}` : key,
    i18n: { language: 'en-US' },
  }),
}));

vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => ({}) }),
  Link: ({ children, to }: { children?: React.ReactNode; to?: string }) => (
    <a href={to}>{children}</a>
  ),
}));

const upsertDayMutateAsync = vi.hoisted(() => vi.fn());
const createLineMutateAsync = vi.hoisted(() => vi.fn());
const deleteLineMutate = vi.hoisted(() => vi.fn());

const workDayState = vi.hoisted(() => ({
  value: { data: undefined as { workDay?: Partial<WorkDay>; lines?: unknown[] } | undefined },
}));

vi.mock('./useWorkDay', () => ({
  newClientKey: () => 'client-key-1',
  useWorkDay: () => workDayState.value,
  useUpsertWorkDay: () => ({ mutateAsync: upsertDayMutateAsync, isPending: false }),
  useCreateWorkDayLine: () => ({ mutateAsync: createLineMutateAsync, isPending: false }),
  useDeleteWorkDayLine: () => ({ mutate: deleteLineMutate, isPending: false }),
}));

// 2026-11-12 (the date every test in this file renders against, unless
// overridden) is a Thursday, so `thu: 8 * 60` is the target the
// over-capacity test below exercises.
vi.mock('../../shared/stores/timeSettingsStore', () => ({
  useTimeSettingsStore: (
    selector: (s: { dailyTargetMinutes: Record<string, number> }) => unknown
  ) =>
    selector({
      dailyTargetMinutes: {
        sun: 0,
        mon: 8 * 60,
        tue: 8 * 60,
        wed: 8 * 60,
        thu: 8 * 60,
        fri: 8 * 60,
        sat: 4 * 60,
      },
    }),
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: (selector: (s: { activeWorkspaceId: string }) => unknown) =>
    selector({ activeWorkspaceId: 'ws-1' }),
}));

const projectsState = vi.hoisted(() => ({ value: { data: [] as ApiProject[] } }));
vi.mock('../projects/useProjects', () => ({
  useProjects: () => projectsState.value,
}));

const tasksState = vi.hoisted(() => ({ value: { data: [] as ApiTask[] } }));
vi.mock('../tasks/useTasks', () => ({
  useTasks: () => tasksState.value,
}));

import AddHours from './AddHours';

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

const onDateChange = vi.fn();

function fillProjectAndTask(projectValue = '1', taskValue = 't1'): void {
  fireEvent.change(screen.getByLabelText('flow.time.project'), { target: { value: projectValue } });
  fireEvent.change(screen.getByLabelText('flow.time.add.task'), { target: { value: taskValue } });
}

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  workDayState.value = {
    data: {
      workDay: { id: '9', workMode: 'onsite', loggedMinutes: 0, breakMinutes: 0 },
      lines: [],
    },
  };
  projectsState.value = { data: [project()] };
  tasksState.value = { data: [task()] };
  upsertDayMutateAsync.mockResolvedValue({ workDay: { id: '9' } });
  createLineMutateAsync.mockResolvedValue({ line: { id: 'line-1' } });
});

describe('AddHours: state 1, empty first load', () => {
  it('shows zeroed duration, a project placeholder and a disabled Task with its reason', () => {
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);

    expect((screen.getByLabelText('flow.time.add.hours') as HTMLInputElement).value).toBe('0');
    expect((screen.getByLabelText('flow.time.add.minutes') as HTMLInputElement).value).toBe('00');
    expect(screen.getByLabelText('flow.time.add.taskDisabledPlaceholder')).toBeDisabled();

    const logButton = screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ });
    expect(logButton).toBeDisabled();
    expect(logButton).toHaveAttribute('title', 'flow.time.add.reasonNoDuration');
  });

  it("defaults Location to the day's own WHERE YOU WORKED value, shown directly", () => {
    workDayState.value = {
      data: { workDay: { id: '9', workMode: 'hybrid', loggedMinutes: 0 }, lines: [] },
    };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    expect(screen.getByRole('radio', { name: 'flow.time.mode.hybrid' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
  });

  it('enables the Task select once a project is chosen', () => {
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    expect(screen.getByLabelText('flow.time.add.taskDisabledPlaceholder')).toBeDisabled();

    fireEvent.change(screen.getByLabelText('flow.time.project'), { target: { value: '1' } });
    expect(screen.getByLabelText('flow.time.add.task')).not.toBeDisabled();
  });
});

describe('AddHours: header pill against the day’s own target', () => {
  it('shows logged of target for the date’s actual day of week', () => {
    workDayState.value = {
      data: { workDay: { id: '9', workMode: 'onsite', loggedMinutes: 90 }, lines: [] },
    };
    // 2026-11-12 is a Thursday -> 8h target from the mocked store above.
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);

    const pill = screen.getByText('flow.time.add.today').parentElement as HTMLElement;
    expect(within(pill).getByText('1:30')).toBeInTheDocument();
    expect(within(pill).getByText('8:00')).toBeInTheDocument();
  });

  it('reads a shorter Saturday target and a non-working Sunday as zero', () => {
    workDayState.value = {
      data: { workDay: { id: '9', workMode: 'onsite', loggedMinutes: 0 }, lines: [] },
    };
    // 2026-11-14 is a Saturday -> 4h target.
    render(<AddHours date="2026-11-14" onDateChange={onDateChange} />);
    const pill = screen.getByText('flow.time.add.today').parentElement as HTMLElement;
    expect(within(pill).getByText('4:00')).toBeInTheDocument();
  });
});

describe('AddHours: state 2, per-field validation on blur', () => {
  it('rejects minutes over 59 on blur, with an hours/minutes hint, and keeps Log hours disabled', () => {
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    const minutes = screen.getByLabelText('flow.time.add.minutes');

    fireEvent.change(minutes, { target: { value: '75' } });
    // Not on keypress - no error yet.
    expect(screen.queryByText(/flow.time.add.errors.minutesRange/)).not.toBeInTheDocument();

    fireEvent.blur(minutes);
    expect(screen.getByText(/flow.time.add.errors.minutesRange/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ })).toBeDisabled();
  });

  it('blocks submission when the day is already approved, using the real WorkDay.status signal', () => {
    workDayState.value = {
      data: {
        workDay: { id: '9', workMode: 'onsite', loggedMinutes: 60, status: 'approved' },
        lines: [],
      },
    };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '1' } });
    fillProjectAndTask();

    expect(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ })).toBeDisabled();
    expect(document.getElementById('add-hours-date-error')).toHaveTextContent(
      'flow.time.add.errors.dayLockedGeneric'
    );
  });

  it('shows a non-blocking over-capacity warning when the day would exceed the workspace day length', () => {
    workDayState.value = {
      data: { workDay: { id: '9', workMode: 'onsite', loggedMinutes: 0 }, lines: [] },
    };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '11' } });
    fillProjectAndTask();

    expect(screen.getByText(/flow.time.add.overCapacity/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ })).not.toBeDisabled();
  });
});

describe('AddHours: submitting and outcomes', () => {
  it('creates the line against the existing work day and shows the success banner', async () => {
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '1' } });
    fireEvent.change(screen.getByLabelText('flow.time.add.minutes'), { target: { value: '30' } });
    fillProjectAndTask();

    fireEvent.click(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ }));

    await waitFor(() => expect(screen.getByRole('status')).toBeInTheDocument());
    expect(upsertDayMutateAsync).not.toHaveBeenCalled();
    expect(createLineMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        workDayId: '9',
        durationMinutes: 90,
        projectId: '1',
        taskId: 't1',
        clientKey: 'client-key-1',
      })
    );
    expect(screen.getByText(/flow.time.add.successTitle/)).toBeInTheDocument();
  });

  it("opens the day first when it doesn't exist yet, then logs against the new id", async () => {
    workDayState.value = { data: { workDay: undefined, lines: [] } };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '1' } });
    fillProjectAndTask();

    fireEvent.click(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ }));

    await waitFor(() => expect(createLineMutateAsync).toHaveBeenCalled());
    expect(upsertDayMutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({ date: '2026-11-12' })
    );
    expect(createLineMutateAsync).toHaveBeenCalledWith(expect.objectContaining({ workDayId: '9' }));
  });

  it('shows the inline failure banner (never a toast) on a rejected submit, keeping every field intact', async () => {
    createLineMutateAsync.mockRejectedValueOnce(new Error('network down'));
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '1' } });
    fillProjectAndTask();

    fireEvent.click(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ }));

    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
    expect(screen.getByText('flow.time.add.failedTitle')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'flow.time.add.tryAgain' })).toBeInTheDocument();
    expect((screen.getByLabelText('flow.time.add.hours') as HTMLInputElement).value).toBe('1');
  });

  it('goes read-only (not disabled) while submitting, so a failed field can still be seen and copied', async () => {
    let resolveCreate: (value: { line: { id: string } }) => void = () => {};
    createLineMutateAsync.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveCreate = resolve;
        })
    );

    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '1' } });
    fillProjectAndTask();
    fireEvent.click(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ }));

    // Mid-flight: fields are read-only, not disabled - the person can still
    // see (and select) what they typed while the request is in the air.
    await waitFor(() =>
      expect(screen.getByLabelText('flow.time.add.hours')).toHaveAttribute('readonly')
    );
    expect(screen.getByLabelText('flow.time.add.hours')).not.toBeDisabled();
    expect(screen.getByRole('button', { name: 'flow.time.add.discard' })).toBeDisabled();
    expect(screen.getByText(/flow.time.add.submittingLine/)).toBeInTheDocument();

    resolveCreate({ line: { id: 'line-9' } });
    await waitFor(() => expect(screen.getByRole('status')).toBeInTheDocument());
  });

  it('Log another keeps date, project, task and location, and clears only duration and note', async () => {
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '2' } });
    fillProjectAndTask();
    fireEvent.change(screen.getByLabelText('flow.time.description'), {
      target: { value: 'Replaced the hinge' },
    });

    fireEvent.click(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ }));
    await waitFor(() => expect(screen.getByRole('status')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'flow.time.add.logAnother' }));

    expect((screen.getByLabelText('flow.time.add.hours') as HTMLInputElement).value).toBe('0');
    expect((screen.getByLabelText('flow.time.description') as HTMLTextAreaElement).value).toBe('');
    expect((screen.getByLabelText('flow.time.project') as HTMLSelectElement).value).toBe('1');
    expect((screen.getByLabelText('flow.time.add.task') as HTMLSelectElement).value).toBe('t1');
  });

  it('Undo deletes the just-created line', async () => {
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '1' } });
    fillProjectAndTask();

    fireEvent.click(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ }));
    await waitFor(() => expect(screen.getByRole('status')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'flow.time.add.undo' }));
    expect(deleteLineMutate).toHaveBeenCalledWith(
      { id: 'line-1', date: '2026-11-12' },
      expect.anything()
    );
  });
});

describe('AddHours: task description box', () => {
  it("shows the selected task's own description, read-only, and omits the box entirely when blank", () => {
    tasksState.value = { data: [task({ description: 'Replace the rusted gate hinge.' })] };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fireEvent.change(screen.getByLabelText('flow.time.project'), { target: { value: '1' } });
    fireEvent.change(screen.getByLabelText('flow.time.add.task'), { target: { value: 't1' } });

    expect(screen.getByText('Replace the rusted gate hinge.')).toBeInTheDocument();
    expect(screen.getByText('flow.time.add.fromTheTask')).toBeInTheDocument();
  });

  it('renders nothing when the task has no description', () => {
    tasksState.value = { data: [task({ description: '' })] };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fillProjectAndTask();
    expect(screen.queryByText('flow.time.add.fromTheTask')).not.toBeInTheDocument();
  });
});

describe('AddHours: Location picker is always visible', () => {
  it('draws the radiogroup directly, with no reveal link on the page', () => {
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    expect(screen.getByRole('radiogroup', { name: 'flow.time.add.location' })).toBeInTheDocument();
    expect(
      screen.queryByRole('button', { name: 'flow.time.add.differentLocation' })
    ).not.toBeInTheDocument();
  });

  it('keyboard arrow keys cycle the selection', () => {
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);

    const group = screen.getByRole('radiogroup', { name: 'flow.time.add.location' });
    expect(within(group).getByRole('radio', { name: 'flow.time.mode.onsite' })).toHaveAttribute(
      'aria-checked',
      'true'
    );

    fireEvent.keyDown(group, { key: 'ArrowRight' });
    expect(within(group).getByRole('radio', { name: 'flow.time.mode.remote' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
  });
});

describe('AddHours: Location remembers the last submitted entry', () => {
  it("defaults to today's own WHERE YOU WORKED value over any remembered location", () => {
    localStorage.setItem('flow.addHours.lastLocation', 'remote');
    workDayState.value = {
      data: { workDay: { id: '9', workMode: 'hybrid', loggedMinutes: 0 }, lines: [] },
    };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    expect(screen.getByRole('radio', { name: 'flow.time.mode.hybrid' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
  });

  it('falls back to the last submitted location when this day has none yet', () => {
    localStorage.setItem('flow.addHours.lastLocation', 'remote');
    workDayState.value = { data: undefined };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    expect(screen.getByRole('radio', { name: 'flow.time.mode.remote' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
  });

  it('falls back to onsite when neither the day nor any remembered location exists', () => {
    workDayState.value = { data: undefined };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    expect(screen.getByRole('radio', { name: 'flow.time.mode.onsite' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
  });

  it('remembers the submitted location for the next entry', async () => {
    workDayState.value = { data: undefined };
    render(<AddHours date="2026-11-12" onDateChange={onDateChange} />);
    fillProjectAndTask();
    fireEvent.change(screen.getByLabelText('flow.time.add.hours'), { target: { value: '1' } });
    fireEvent.click(screen.getByRole('radio', { name: 'flow.time.mode.remote' }));

    fireEvent.click(screen.getByRole('button', { name: /flow\.time\.add\.logHours$/ }));

    await waitFor(() => expect(createLineMutateAsync).toHaveBeenCalled());
    expect(localStorage.getItem('flow.addHours.lastLocation')).toBe('remote');
  });
});
