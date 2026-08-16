import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { LeaveTypePayload } from './useLeave';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en-US' } }),
}));

const createRequestMutate = vi.hoisted(() => vi.fn());
vi.mock('./useLeave', () => ({
  useCreateLeaveRequest: () => ({ mutate: createRequestMutate, isPending: false }),
}));

import RequestLeaveDialog from './RequestLeaveDialog';

function leaveType(over: Partial<LeaveTypePayload> = {}): LeaveTypePayload {
  return { id: '1', name: 'Annual Leave', isPaid: true, requiresApproval: true, ...over };
}

const defaultProps = {
  open: true,
  onOpenChange: vi.fn(),
  employeeId: '4201',
  leaveTypes: [leaveType()],
};

beforeEach(() => {
  vi.clearAllMocks();
});

describe('RequestLeaveDialog', () => {
  it('renders the configured leave types in the picker', () => {
    render(
      <RequestLeaveDialog
        {...defaultProps}
        leaveTypes={[
          leaveType({ id: '1', name: 'Annual Leave' }),
          leaveType({ id: '2', name: 'Sick Leave' }),
        ]}
      />
    );
    const select = screen.getByLabelText('flow.leave.leaveType') as HTMLSelectElement;
    expect(Array.from(select.options).map((o) => o.textContent)).toEqual([
      'flow.leave.selectLeaveType',
      'Annual Leave',
      'Sick Leave',
    ]);
  });

  it('keeps submit disabled until type and both dates are set', () => {
    render(<RequestLeaveDialog {...defaultProps} />);
    expect(screen.getByRole('button', { name: 'flow.leave.submit' })).toBeDisabled();
  });

  it("submits the caller's own employeeId, the picked type, and computed total days", () => {
    render(<RequestLeaveDialog {...defaultProps} />);

    fireEvent.change(screen.getByLabelText('flow.leave.leaveType'), { target: { value: '1' } });
    fireEvent.change(screen.getByLabelText('flow.leave.startDate'), {
      target: { value: '2027-01-04' },
    });
    fireEvent.change(screen.getByLabelText('flow.leave.endDate'), {
      target: { value: '2027-01-06' },
    });

    const submit = screen.getByRole('button', { name: 'flow.leave.submit' });
    expect(submit).not.toBeDisabled();
    fireEvent.click(submit);

    expect(createRequestMutate).toHaveBeenCalledWith(
      expect.objectContaining({
        employeeId: '4201',
        leaveTypeId: '1',
        startDate: '2027-01-04',
        endDate: '2027-01-06',
        totalDays: 3,
      }),
      expect.anything()
    );
  });
});
