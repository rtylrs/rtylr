import { describe, expect, it } from 'vitest';
import { getLeaveRequestScope } from './leaveAccess';

describe('getLeaveRequestScope', () => {
  it('requests only the current employee leave for ordinary users', () => {
    expect(getLeaveRequestScope('employee-42', false)).toEqual({
      employeeId: 'employee-42',
      status: undefined,
    });
  });

  it('leaves the employee filter open for approvers to receive the team feed', () => {
    expect(getLeaveRequestScope('approver-7', true)).toEqual({
      employeeId: undefined,
      status: undefined,
    });
  });

  it('does not send an empty employee filter while the session is hydrating', () => {
    expect(getLeaveRequestScope('', false)).toEqual({
      employeeId: undefined,
      status: undefined,
    });
  });
});
