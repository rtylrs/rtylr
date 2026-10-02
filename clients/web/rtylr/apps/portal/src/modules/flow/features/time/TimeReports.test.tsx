import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import userEvent from '@testing-library/user-event';
import type { FlowTeamMember } from '../../shared/hooks/queries/useFlowTeam';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

const navigateSpy = vi.hoisted(() => vi.fn());

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateSpy,
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: (selector: (s: { activeWorkspaceId: string }) => unknown) =>
    selector({ activeWorkspaceId: 'ws-1' }),
}));

const timeSettingsState = vi.hoisted(() => ({
  value: { weekStart: 'monday' as 'monday' | 'sunday', trackAttendance: true },
}));

vi.mock('../../shared/stores/timeSettingsStore', () => ({
  useTimeSettingsStore: (selector: (s: typeof timeSettingsState.value) => unknown) =>
    selector(timeSettingsState.value),
}));

const membersState = vi.hoisted(() => ({
  value: { data: [] as FlowTeamMember[], isLoading: false },
}));

vi.mock('../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => membersState.value,
}));

const workDaysState = vi.hoisted(() => ({
  value: {
    data: undefined as { workDays: Array<Record<string, unknown>> } | undefined,
    isLoading: false,
    isError: false,
  },
}));

vi.mock('./useWorkDay', () => ({
  todayLocalDate: () => '2026-08-16',
  useTeamWorkDays: () => workDaysState.value,
}));

import TimeReports from './TimeReports';

// A FlowTeamMember's `id` is the user id, which is what the totals are keyed on.
function member(over: Partial<FlowTeamMember> = {}): FlowTeamMember {
  return {
    id: '7',
    name: 'Maya Krishnan',
    email: 'maya@example.test',
    memberType: 'member',
    isOwner: false,
    hasFlowAccess: true,
    isActive: true,
    ...over,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  timeSettingsState.value = { weekStart: 'monday', trackAttendance: true };
  membersState.value = {
    data: [member(), member({ id: '9', name: 'Sam Okafor' })],
    isLoading: false,
  };
  workDaysState.value = {
    data: {
      workDays: [
        { id: '1', userId: '7', clockedMinutes: 480, loggedMinutes: 450, gapMinutes: 30 },
        { id: '2', userId: '7', clockedMinutes: 240, loggedMinutes: 240, gapMinutes: 0 },
        { id: '3', userId: '9', clockedMinutes: 480, loggedMinutes: 400, gapMinutes: 80 },
      ],
    },
    isLoading: false,
    isError: false,
  };
});

describe('TimeReports', () => {
  // These were skipped while the member source was a hardcoded `[]`, with a note
  // to un-skip once useFlowTeam landed. It has, so they run again. The
  // assertions are unchanged from when they were written.
  it('sums clocked, logged and gap minutes per person across the range', () => {
    render(<TimeReports />);

    const mayaRow = screen.getByText('Maya Krishnan').closest('tr');
    expect(mayaRow?.textContent).toContain('12:00'); // 480 + 240 clocked
    expect(mayaRow?.textContent).toContain('11:30'); // 450 + 240 logged
    expect(mayaRow?.textContent).toContain('0:30'); // 30 + 0 gap

    const samRow = screen.getByText('Sam Okafor').closest('tr');
    expect(samRow?.textContent).toContain('8:00'); // 480 clocked
    expect(samRow?.textContent).toContain('1:20'); // 80 gap
  });

  it('shows a member with no days in the range as all zero rather than omitting them', () => {
    membersState.value = {
      data: [member(), member({ id: '11', name: 'No Entries Yet' })],
      isLoading: false,
    };
    render(<TimeReports />);
    const row = screen.getByText('No Entries Yet').closest('tr');
    expect(row?.textContent).toContain('0:00');
  });

  it('renders a permission-denied message instead of crashing when the caller lacks access', () => {
    workDaysState.value = { data: undefined, isLoading: false, isError: true };
    render(<TimeReports />);
    expect(screen.getByText('flow.time.teamAccessDenied')).toBeInTheDocument();
  });

  it('shows a real empty state instead of a headers-only table when nobody has Flow access', () => {
    membersState.value = { data: [], isLoading: false };
    render(<TimeReports />);
    expect(screen.getByText('flow.time.reportsEmptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.time.reportsEmptyDescription')).toBeInTheDocument();
  });

  // useFlowTeam returns the whole business team, not just Flow's users, so the
  // filter is what keeps someone who has no Flow access out of a Flow report.
  it('leaves out a business member who does not have Flow access', () => {
    membersState.value = {
      data: [member(), member({ id: '9', name: 'Dana Reeves', hasFlowAccess: false })],
      isLoading: false,
    };
    render(<TimeReports />);
    expect(screen.getByText('Maya Krishnan')).toBeInTheDocument();
    expect(screen.queryByText('Dana Reeves')).not.toBeInTheDocument();
  });

  // The empty state's CTA used to be `onClick: () => {}`. A button that does
  // nothing is worse than no button, so this pins it to a real destination.
  it('sends the empty-state CTA to the settings screen that grants Flow access', async () => {
    membersState.value = { data: [], isLoading: false };
    render(<TimeReports />);
    await userEvent.click(screen.getByRole('button', { name: 'flow.time.reportsEmptyCta' }));
    expect(navigateSpy).toHaveBeenCalledWith({ to: '/settings/flow', search: { section: 'team' } });
  });
});

describe('TimeReports: trackAttendance gating', () => {
  it('shows Clocked and Unaccounted columns by default', () => {
    render(<TimeReports />);
    expect(screen.getByText('flow.time.clocked')).toBeInTheDocument();
    expect(screen.getByText('flow.time.unaccounted')).toBeInTheDocument();
    expect(screen.getByText('flow.time.logged')).toBeInTheDocument();
  });

  it('hides Clocked and Unaccounted, keeping only Logged, when attendance tracking is off', () => {
    timeSettingsState.value = { ...timeSettingsState.value, trackAttendance: false };
    render(<TimeReports />);
    expect(screen.queryByText('flow.time.clocked')).not.toBeInTheDocument();
    expect(screen.queryByText('flow.time.unaccounted')).not.toBeInTheDocument();
    expect(screen.getByText('flow.time.logged')).toBeInTheDocument();

    const mayaRow = screen.getByText('Maya Krishnan').closest('tr');
    // 450 + 240 logged, with no clocked/gap columns to conflict with it.
    expect(mayaRow?.textContent).toContain('11:30');
  });
});
