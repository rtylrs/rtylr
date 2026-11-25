import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { initTestI18n } from '../../shared/testing/i18n';

const mocks = vi.hoisted(() => ({ openTask: vi.fn(), args: [] as unknown[] }));

vi.mock('../../shared/lib/useOpenTask', () => ({ useOpenTask: () => mocks.openTask }));
vi.mock('../../shared/lib/dateUtils', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../shared/lib/dateUtils')>()),
  todayIso: () => '2026-11-25',
}));
vi.mock('./useWorkDay', () => ({
  usePlannedWorkload: (...args: unknown[]) => {
    mocks.args = args;
    return {
      isPending: false,
      isError: false,
      data: {
        everyone: false,
        people: ['7', '8'],
        days: [{ userId: '7', date: '2026-11-25', plannedMinutes: 450, taskIds: ['41'] }],
        tasks: [
          {
            id: '41',
            identifier: 'T-41',
            title: 'IH screen',
            estimateMinutes: 450,
            startDate: '',
            dueDate: '2026-11-25',
          },
        ],
      },
    };
  },
}));

import PlannedWorkload from './PlannedWorkload';

beforeEach(async () => {
  vi.clearAllMocks();
  localStorage.clear();
  await initTestI18n();
});

describe('PlannedWorkload', () => {
  const team = [
    { id: '7', name: 'Zed' },
    { id: '8', name: 'Amal' },
    { id: '9', name: 'Not visible' },
  ];

  it('asks for two weeks from Monday and shows only the people the server allows', () => {
    render(<PlannedWorkload team={team} />);
    expect(mocks.args).toEqual(['2026-11-23', '2026-12-06', '1,2,3,4,5']);
    expect(screen.getByRole('rowheader', { name: 'Zed' })).toBeInTheDocument();
    expect(screen.getByRole('rowheader', { name: 'Amal' })).toBeInTheDocument();
    expect(screen.queryByText('Not visible')).not.toBeInTheDocument();
  });

  it('opens the tasks behind a cell', () => {
    render(<PlannedWorkload team={team} />);
    fireEvent.click(screen.getByRole('button', { name: /Zed, .*7\.5 hours planned/ }));
    fireEvent.click(screen.getByRole('button', { name: /T-41/ }));
    expect(mocks.openTask).toHaveBeenCalledWith('41');
  });

  it('remembers a Sunday work week and asks the server for it', () => {
    localStorage.setItem(
      'flow.teamHours.planned',
      JSON.stringify({ workWeek: 'sun_thu', targetHours: 6 })
    );
    render(<PlannedWorkload team={team} />);
    expect(mocks.args).toEqual(['2026-11-22', '2026-12-05', '7,1,2,3,4']);
  });
});
