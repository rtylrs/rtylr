import { describe, it, expect } from 'vitest';
import {
  employmentStatusLabel,
  employmentStatusColor,
  isActiveEmployee,
  isOnLeave,
  isTerminated,
  contractTypeLabel,
  isPermanentEmployee,
  isTemporaryWorker,
  headcountByStatus,
} from './hrStatusUtils';

describe('employmentStatusLabel', () => {
  it('returns label for Active', () => {
    expect(employmentStatusLabel('Active')).toBe('Active');
  });
  it('returns label for On Leave', () => {
    expect(employmentStatusLabel('On Leave')).toBe('On Leave');
  });
  it('returns label for Terminated', () => {
    expect(employmentStatusLabel('Terminated')).toBe('Terminated');
  });
});

describe('employmentStatusColor', () => {
  it('returns green for Active', () => {
    expect(employmentStatusColor('Active')).toBe('green');
  });
  it('returns red for Terminated', () => {
    expect(employmentStatusColor('Terminated')).toBe('red');
  });
  it('returns yellow for On Leave', () => {
    expect(employmentStatusColor('On Leave')).toBe('yellow');
  });
});

describe('isActiveEmployee', () => {
  it('returns true for Active', () => {
    expect(isActiveEmployee('Active')).toBe(true);
  });
  it('returns true for Probation', () => {
    expect(isActiveEmployee('Probation')).toBe(true);
  });
  it('returns false for Terminated', () => {
    expect(isActiveEmployee('Terminated')).toBe(false);
  });
});

describe('isOnLeave', () => {
  it('returns true for On Leave', () => {
    expect(isOnLeave('On Leave')).toBe(true);
  });
  it('returns false for Active', () => {
    expect(isOnLeave('Active')).toBe(false);
  });
});

describe('isTerminated', () => {
  it('returns true for Terminated', () => {
    expect(isTerminated('Terminated')).toBe(true);
  });
  it('returns false for Active', () => {
    expect(isTerminated('Active')).toBe(false);
  });
});

describe('contractTypeLabel', () => {
  it('returns label for Full-time', () => {
    expect(contractTypeLabel('Full-time')).toBe('Full-time');
  });
  it('returns label for Intern', () => {
    expect(contractTypeLabel('Intern')).toBe('Intern');
  });
});

describe('isPermanentEmployee', () => {
  it('returns true for Full-time', () => {
    expect(isPermanentEmployee('Full-time')).toBe(true);
  });
  it('returns true for Part-time', () => {
    expect(isPermanentEmployee('Part-time')).toBe(true);
  });
  it('returns false for Contract', () => {
    expect(isPermanentEmployee('Contract')).toBe(false);
  });
});

describe('isTemporaryWorker', () => {
  it('returns true for Contract', () => {
    expect(isTemporaryWorker('Contract')).toBe(true);
  });
  it('returns true for Intern', () => {
    expect(isTemporaryWorker('Intern')).toBe(true);
  });
  it('returns false for Full-time', () => {
    expect(isTemporaryWorker('Full-time')).toBe(false);
  });
});

describe('headcountByStatus', () => {
  it('counts employees by status', () => {
    const employees = [
      { status: 'Active' as const },
      { status: 'Active' as const },
      { status: 'Terminated' as const },
    ];
    const result = headcountByStatus(employees);
    expect(result.Active).toBe(2);
    expect(result.Terminated).toBe(1);
    expect(result.Probation).toBe(0);
  });
});
