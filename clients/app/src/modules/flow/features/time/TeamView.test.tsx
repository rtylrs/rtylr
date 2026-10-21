import React from 'react';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { FlowTeamMember } from '../../shared/hooks/queries/useFlowTeam';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => opts?.reason ?? key,
    i18n: { language: 'en-US' },
  }),
}));

vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: (selector: (s: { activeWorkspaceId: string }) => unknown) =>
    selector({ activeWorkspaceId: 'ws-1' }),
}));

const navigateSpy = vi.hoisted(() => vi.fn());

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateSpy,
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
    error: undefined as unknown,
  },
}));

vi.mock('./useWorkDay', () => ({
  todayLocalDate: () => '2026-08-16',
  useTeamWorkDays: () => workDaysState.value,
}));

import TeamView from './TeamView';

// A FlowTeamMember's `id` is the user id, which is what a workDay is keyed on.
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
  membersState.value = {
    data: [member(), member({ id: '9', name: 'Sam Okafor' })],
    isLoading: false,
  };
  workDaysState.value = {
    data: {
      workDays: [{ id: '1', userId: '7', clockedMinutes: 480, loggedMinutes: 465, gapMinutes: 15 }],
    },
    isLoading: false,
    isError: false,
    error: undefined,
  };
});

describe('TeamView', () => {
  // These two were skipped while the member source was a hardcoded `[]`, with a
  // note to un-skip once useFlowTeam landed. It has, so they run again. The
  // assertions are unchanged from when they were written.
  it('lists every member with Flow access, matching each to their day by user id', () => {
    render(<TeamView />);
    expect(screen.getByText('Maya Krishnan')).toBeInTheDocument();
    expect(screen.getByText('Sam Okafor')).toBeInTheDocument();
  });

  it('shows a not-started state for a member with no recorded day', () => {
    render(<TeamView />);
    // Sam (user id 9) has no matching workDay in the fixture above.
    const samRow = screen.getByText('Sam Okafor').closest('li');
    expect(samRow).not.toBeNull();
    expect(samRow?.textContent).toContain('flow.time.teamNotStarted');
  });

  // useFlowTeam returns the whole business team, not just Flow's users, so the
  // filter is what keeps someone who has no Flow access off a Flow time sheet.
  it('leaves out a business member who does not have Flow access', () => {
    membersState.value = {
      data: [member(), member({ id: '9', name: 'Dana Reeves', hasFlowAccess: false })],
      isLoading: false,
    };
    render(<TeamView />);
    expect(screen.getByText('Maya Krishnan')).toBeInTheDocument();
    expect(screen.queryByText('Dana Reeves')).not.toBeInTheDocument();
  });

  it('renders a permission-denied message instead of crashing when the caller lacks access', () => {
    workDaysState.value = {
      data: undefined,
      isLoading: false,
      isError: true,
      error: { statusCode: 403 },
    };
    render(<TeamView />);
    expect(screen.getByText('flow.time.teamAccessDenied')).toBeInTheDocument();
  });

  it('shows a real empty state instead of a blank list when nobody has Flow access', () => {
    membersState.value = { data: [], isLoading: false };
    render(<TeamView />);
    expect(screen.getByText('flow.time.teamEmptyTitle')).toBeInTheDocument();
    expect(screen.getByText('flow.time.teamEmptyDescription')).toBeInTheDocument();
  });

  // The empty state's CTA used to be `onClick: () => {}`. A button that does
  // nothing is worse than no button, so this pins it to a real destination.
  it('sends the empty-state CTA to the settings screen that grants Flow access', async () => {
    membersState.value = { data: [], isLoading: false };
    render(<TeamView />);
    await userEvent.click(screen.getByRole('button', { name: 'flow.time.teamEmptyCta' }));
    expect(navigateSpy).toHaveBeenCalledWith({ to: '/team' });
  });
});
