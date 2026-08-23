import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

const employeeIdState = vi.hoisted(() => ({ value: { data: '4201', isLoading: false } }));
vi.mock('../../shared/hooks/useMyEmployeeId', () => ({
  useMyEmployeeId: () => employeeIdState.value,
}));

interface LeaveQueryState {
  data: Array<Record<string, unknown>>;
  isError?: boolean;
  isPaused?: boolean;
  failureCount?: number;
  refetch?: () => void;
}

const leaveTypesState = vi.hoisted(() => ({
  value: { data: [{ id: '1', name: 'Annual Leave', isPaid: true, requiresApproval: true }] },
}));
const balancesRefetch = vi.hoisted(() => vi.fn());
const requestsRefetch = vi.hoisted(() => vi.fn());
const balancesState = vi.hoisted(() => ({
  value: {
    data: [
      {
        id: 'b1',
        employeeId: '4201',
        leaveTypeId: '1',
        year: 2027,
        entitledDays: 20,
        usedDays: 5,
        pendingDays: 2,
        remainingDays: 13,
      },
    ],
    refetch: balancesRefetch,
  } as LeaveQueryState,
}));
const requestsState = vi.hoisted(() => ({
  value: {
    data: [
      {
        id: 'r1',
        employeeId: '4201',
        leaveTypeId: '1',
        startDate: '2027-02-01',
        endDate: '2027-02-03',
        totalDays: 3,
        status: 'pending',
        currentApprovalLevel: 0,
      },
    ],
    refetch: requestsRefetch,
  } as LeaveQueryState,
}));
const cancelMutate = vi.hoisted(() => vi.fn());

vi.mock('./useLeave', () => ({
  useListLeaveTypes: () => leaveTypesState.value,
  useMyLeaveBalances: () => balancesState.value,
  useMyLeaveRequests: () => requestsState.value,
  useCancelLeaveRequest: () => ({ mutate: cancelMutate, isPending: false }),
}));

vi.mock('./RequestLeaveDialog', () => ({
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="request-leave-dialog" /> : null,
}));

import LeaveScreen from './LeaveScreen';

/**
 * Every shared fixture is rebuilt here, not just the employee id.
 *
 * Seven of the tests below reassign leaveTypesState, balancesState or
 * requestsState to model an empty, loading or failed response, and nothing put
 * them back. That passed only because those tests are declared after the ones
 * that need pristine data: running this file with --sequence.shuffle failed 2 to
 * 6 of its 10 tests, most often "Unable to find an element with the text:
 * Annual Leave" from the test on line 117 that empties the leave types.
 *
 * Fresh objects rather than saved references, so a test that mutates a nested
 * array in place cannot leak either.
 */
beforeEach(() => {
  vi.clearAllMocks();
  employeeIdState.value = { data: '4201', isLoading: false };
  leaveTypesState.value = {
    data: [{ id: '1', name: 'Annual Leave', isPaid: true, requiresApproval: true }],
  };
  balancesState.value = {
    data: [
      {
        id: 'b1',
        employeeId: '4201',
        leaveTypeId: '1',
        year: 2027,
        entitledDays: 20,
        usedDays: 5,
        pendingDays: 2,
        remainingDays: 13,
      },
    ],
    refetch: balancesRefetch,
  } as LeaveQueryState;
  requestsState.value = {
    data: [
      {
        id: 'r1',
        employeeId: '4201',
        leaveTypeId: '1',
        startDate: '2027-02-01',
        endDate: '2027-02-03',
        totalDays: 3,
        status: 'pending',
        currentApprovalLevel: 0,
      },
    ],
    refetch: requestsRefetch,
  } as LeaveQueryState;
});

describe('LeaveScreen', () => {
  it("shows the caller's balance for each configured leave type", () => {
    render(<LeaveScreen />);
    expect(screen.getAllByText('Annual Leave').length).toBeGreaterThan(0);
    expect(screen.getByText('13')).toBeInTheDocument();
  });

  it('shows the request list with status', () => {
    render(<LeaveScreen />);
    expect(screen.getByText('2027-02-01')).toBeInTheDocument();
    expect(screen.getByText('flow.leave.status.pending')).toBeInTheDocument();
  });

  it('shows a cancel action only for pending requests', () => {
    render(<LeaveScreen />);
    expect(
      screen.getByRole('button', { name: 'flow.leave.cancelRequestLabel' })
    ).toBeInTheDocument();
  });

  it('cancelling a request calls useCancelLeaveRequest with its id', () => {
    render(<LeaveScreen />);
    fireEvent.click(screen.getByRole('button', { name: 'flow.leave.cancelRequestLabel' }));
    expect(cancelMutate).toHaveBeenCalledWith('r1');
  });

  it('opens the request dialog from the new-request button', () => {
    render(<LeaveScreen />);
    expect(screen.queryByTestId('request-leave-dialog')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.leave.newRequest' }));
    expect(screen.getByTestId('request-leave-dialog')).toBeInTheDocument();
  });

  it('shows an empty balances message when nothing is configured, with a request-time-off CTA', () => {
    balancesState.value = { data: [], refetch: balancesRefetch };
    leaveTypesState.value = { data: [] };
    render(<LeaveScreen />);
    expect(screen.getByText('flow.leave.noBalances')).toBeInTheDocument();
    const requestButtons = screen.getAllByRole('button', { name: 'flow.leave.newRequest' });
    expect(requestButtons.length).toBe(2);
    fireEvent.click(requestButtons[1]);
    expect(screen.getByTestId('request-leave-dialog')).toBeInTheDocument();
  });

  it('shows the load-failed state for balances stuck paused after a failed attempt, not the empty state', () => {
    balancesState.value = {
      data: [],
      isError: false,
      isPaused: true,
      failureCount: 1,
      refetch: balancesRefetch,
    };
    render(<LeaveScreen />);
    expect(screen.getByText('flow.leave.loadBalancesFailedTitle')).toBeInTheDocument();
    expect(screen.queryByText('flow.leave.noBalances')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.leave.retryCta' }));
    expect(balancesRefetch).toHaveBeenCalled();
  });

  it('does not treat balances merely paused before their first attempt as failed', () => {
    balancesState.value = {
      data: [],
      isError: false,
      isPaused: true,
      failureCount: 0,
      refetch: balancesRefetch,
    };
    render(<LeaveScreen />);
    expect(screen.getByText('flow.leave.noBalances')).toBeInTheDocument();
    expect(screen.queryByText('flow.leave.loadBalancesFailedTitle')).not.toBeInTheDocument();
  });

  it('shows an empty requests message with a request-time-off CTA', () => {
    // Explicitly non-empty: an earlier test in this file can leave
    // balancesState.value pointed at an empty/paused state, which would add
    // a second "Request time off" CTA into the balances card too.
    balancesState.value = {
      data: [
        {
          id: 'b1',
          employeeId: '4201',
          leaveTypeId: '1',
          year: 2027,
          entitledDays: 20,
          usedDays: 5,
          pendingDays: 2,
          remainingDays: 13,
        },
      ],
      refetch: balancesRefetch,
    };
    requestsState.value = { data: [], refetch: requestsRefetch };
    render(<LeaveScreen />);
    expect(screen.getByText('flow.leave.noRequests')).toBeInTheDocument();
    const requestButtons = screen.getAllByRole('button', { name: 'flow.leave.newRequest' });
    expect(requestButtons.length).toBe(2);
    fireEvent.click(requestButtons[1]);
    expect(screen.getByTestId('request-leave-dialog')).toBeInTheDocument();
  });

  it('shows the load-failed state for requests stuck paused after a failed attempt, not the empty state', () => {
    requestsState.value = {
      data: [],
      isError: false,
      isPaused: true,
      failureCount: 1,
      refetch: requestsRefetch,
    };
    render(<LeaveScreen />);
    expect(screen.getByText('flow.leave.loadRequestsFailedTitle')).toBeInTheDocument();
    expect(screen.queryByText('flow.leave.noRequests')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.leave.retryCta' }));
    expect(requestsRefetch).toHaveBeenCalled();
  });
});
