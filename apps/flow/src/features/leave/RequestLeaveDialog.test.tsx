import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
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
  vi.useFakeTimers();
  vi.setSystemTime(new Date('2026-08-20T12:00:00.000Z'));
  vi.clearAllMocks();
});

afterEach(() => {
  vi.useRealTimers();
});

// Both pickers default to the pinned "today" above (August 2026), so reaching
// January 2027 always takes exactly 5 "next month" clicks.
function pickLeaveDate(pickerLabel: string, dayAriaLabel: string): void {
  fireEvent.click(screen.getByLabelText(pickerLabel));
  for (let i = 0; i < 5; i++) {
    fireEvent.click(screen.getByRole('button', { name: 'common.datePicker.nextMonth' }));
  }
  fireEvent.click(screen.getByRole('gridcell', { name: dayAriaLabel }));
}

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

  // Without a DialogDescription, Radix warns "Missing Description or
  // aria-describedby" and a screen reader announces the dialog with a title and
  // no statement of what it does. Asserting the key, not English, also catches a
  // description that was added as a hardcoded string.
  it('gives the dialog an accessible description', () => {
    render(<RequestLeaveDialog {...defaultProps} />);
    expect(screen.getByRole('dialog')).toHaveAccessibleDescription(
      'flow.leave.newRequestDescription'
    );
  });

  it('keeps submit disabled until type and both dates are set', () => {
    render(<RequestLeaveDialog {...defaultProps} />);
    expect(screen.getByRole('button', { name: 'flow.leave.submit' })).toBeDisabled();
  });

  it('disables the type field and submit with a reason when no leave types are configured', () => {
    render(<RequestLeaveDialog {...defaultProps} leaveTypes={[]} />);

    const select = screen.getByLabelText('flow.leave.leaveType') as HTMLSelectElement;
    const submit = screen.getByRole('button', { name: 'flow.leave.submit' });

    expect(select).toBeDisabled();
    expect(submit).toBeDisabled();
    expect(submit).toHaveAttribute('title', 'flow.leave.noLeaveTypesReason');

    const describedBy = submit.getAttribute('aria-describedby');
    expect(describedBy).toBeTruthy();
    expect(document.getElementById(describedBy as string)).toHaveTextContent(
      'flow.leave.noLeaveTypesReason'
    );

    // Filling in both dates cannot enable submit either - there is no real
    // type to select, so the gate stays in effect regardless of other fields.
    pickLeaveDate('flow.leave.startDate', 'Monday, January 4, 2027');
    pickLeaveDate('flow.leave.endDate', 'Wednesday, January 6, 2027');
    expect(submit).toBeDisabled();
    expect(createRequestMutate).not.toHaveBeenCalled();
  });

  it('enables submit normally once at least one real leave type exists', () => {
    render(<RequestLeaveDialog {...defaultProps} />);

    const select = screen.getByLabelText('flow.leave.leaveType') as HTMLSelectElement;
    const submit = screen.getByRole('button', { name: 'flow.leave.submit' });

    expect(select).not.toBeDisabled();
    expect(submit).not.toHaveAttribute('title');
    expect(submit).not.toHaveAttribute('aria-describedby');

    fireEvent.change(select, { target: { value: '1' } });
    pickLeaveDate('flow.leave.startDate', 'Monday, January 4, 2027');
    pickLeaveDate('flow.leave.endDate', 'Wednesday, January 6, 2027');

    expect(submit).not.toBeDisabled();
    fireEvent.click(submit);
    expect(createRequestMutate).toHaveBeenCalledTimes(1);
  });

  it("submits the caller's own employeeId, the picked type, and computed total days", () => {
    render(<RequestLeaveDialog {...defaultProps} />);

    fireEvent.change(screen.getByLabelText('flow.leave.leaveType'), { target: { value: '1' } });
    pickLeaveDate('flow.leave.startDate', 'Monday, January 4, 2027');
    pickLeaveDate('flow.leave.endDate', 'Wednesday, January 6, 2027');

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
