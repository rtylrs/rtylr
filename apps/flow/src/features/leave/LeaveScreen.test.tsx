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

const leaveTypesState = vi.hoisted(() => ({
  value: { data: [{ id: '1', name: 'Annual Leave', isPaid: true, requiresApproval: true }] },
}));
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
  },
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
  },
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

beforeEach(() => {
  vi.clearAllMocks();
  employeeIdState.value = { data: '4201', isLoading: false };
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

  it('shows an empty balances message when nothing is configured', () => {
    balancesState.value = { data: [] };
    leaveTypesState.value = { data: [] };
    render(<LeaveScreen />);
    expect(screen.getByText('flow.leave.noBalances')).toBeInTheDocument();
  });
});
