import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiProject } from '../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

const upsertDayMutate = vi.hoisted(() => vi.fn());
const createLineMutate = vi.hoisted(() => vi.fn());
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
  useUpsertWorkDay: () => ({ mutate: upsertDayMutate, isPending: false }),
  useCreateWorkDayLine: () => ({ mutate: createLineMutate, isPending: false }),
  useDeleteWorkDayLine: () => ({ mutate: deleteLineMutate, isPending: false }),
}));

vi.mock('../../shared/stores/timeSettingsStore', () => ({
  useTimeSettingsStore: (selector: (s: { weekStart: 'monday' | 'sunday' }) => unknown) =>
    selector({ weekStart: 'monday' }),
}));

const projectsState = vi.hoisted(() => ({
  value: { data: [] as ApiProject[] },
}));

vi.mock('../projects/useProjects', () => ({
  useProjects: () => projectsState.value,
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
    expect((screen.getByLabelText('flow.time.date') as HTMLInputElement).value).toBe('2026-08-11');
  });
});

describe('MyDay: project on a logged line', () => {
  it('renders a project picker with the workspace projects', () => {
    render(<MyDay />);
    const select = screen.getByLabelText('flow.time.project') as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.textContent)).toEqual([
      'flow.time.selectProject',
      'Site fitout',
      'Warehouse racking',
    ]);
  });

  it('keeps Add disabled until a project is chosen, even with a valid duration', () => {
    render(<MyDay />);
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '1:30' } });
    expect(screen.getByRole('button', { name: 'flow.time.addLine' })).toBeDisabled();
  });

  it('enables Add once duration and project are both set, and sends projectId', () => {
    render(<MyDay />);
    fireEvent.change(screen.getByLabelText('flow.time.duration'), { target: { value: '1:30' } });
    fireEvent.change(screen.getByLabelText('flow.time.project'), { target: { value: '2' } });

    const addButton = screen.getByRole('button', { name: 'flow.time.addLine' });
    expect(addButton).not.toBeDisabled();
    fireEvent.click(addButton);

    expect(createLineMutate).toHaveBeenCalledWith(
      expect.objectContaining({ projectId: '2' }),
      expect.anything()
    );
  });
});
