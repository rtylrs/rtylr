import { describe, it, expect } from 'vitest';
import {
  isBillOverdue,
  getBillBalance,
  isBillPaid,
  getBillStatusLabel,
  deriveBillStatus,
  daysOverdue,
  getBillUrgency,
} from './billUtils';

describe('isBillOverdue', () => {
  it('returns true when due date is past and status is Unpaid', () => {
    expect(isBillOverdue('2020-01-01', 'Unpaid', '2026-06-01')).toBe(true);
  });

  it('returns false when due date is future', () => {
    expect(isBillOverdue('2026-12-31', 'Unpaid', '2026-06-01')).toBe(false);
  });

  it('returns false when status is Paid regardless of date', () => {
    expect(isBillOverdue('2020-01-01', 'Paid', '2026-06-01')).toBe(false);
  });

  it('returns false when dueDate is null', () => {
    expect(isBillOverdue(null, 'Unpaid', '2026-06-01')).toBe(false);
  });
});

describe('getBillBalance', () => {
  it('returns full amount when nothing is paid', () => {
    expect(getBillBalance(1000)).toBe(1000);
  });

  it('returns difference when partially paid', () => {
    expect(getBillBalance(1000, 400)).toBe(600);
  });

  it('returns 0 when fully paid', () => {
    expect(getBillBalance(500, 500)).toBe(0);
  });

  it('handles null amount as 0', () => {
    expect(getBillBalance(null)).toBe(0);
  });
});

describe('isBillPaid', () => {
  it('returns true when fully paid', () => {
    expect(isBillPaid(300, 300)).toBe(true);
  });

  it('returns false when partially paid', () => {
    expect(isBillPaid(300, 150)).toBe(false);
  });

  it('returns false when nothing is paid', () => {
    expect(isBillPaid(500)).toBe(false);
  });
});

describe('getBillStatusLabel', () => {
  it('returns correct label for each status', () => {
    expect(getBillStatusLabel('Unpaid')).toBe('Unpaid');
    expect(getBillStatusLabel('Paid')).toBe('Paid');
    expect(getBillStatusLabel('Overdue')).toBe('Overdue');
    expect(getBillStatusLabel('Partial')).toBe('Partially Paid');
  });
});

describe('deriveBillStatus', () => {
  it('returns Paid when fully paid', () => {
    expect(deriveBillStatus(1000, 1000, null, '2026-06-01')).toBe('Paid');
  });

  it('returns Partial when some is paid', () => {
    expect(deriveBillStatus(1000, 500, '2026-12-31', '2026-06-01')).toBe('Partial');
  });

  it('returns Overdue when nothing paid and past due', () => {
    expect(deriveBillStatus(1000, 0, '2020-01-01', '2026-06-01')).toBe('Overdue');
  });

  it('returns Unpaid when nothing paid and not yet due', () => {
    expect(deriveBillStatus(1000, 0, '2026-12-31', '2026-06-01')).toBe('Unpaid');
  });

  it('returns Unpaid when no due date and nothing paid', () => {
    expect(deriveBillStatus(1000, 0, null, '2026-06-01')).toBe('Unpaid');
  });
});

describe('daysOverdue', () => {
  it('returns 0 when not overdue', () => {
    expect(daysOverdue('2026-12-31', '2026-06-01')).toBe(0);
  });
  it('returns days when overdue', () => {
    expect(daysOverdue('2026-05-22', '2026-06-01')).toBe(10);
  });
  it('returns 0 when null', () => {
    expect(daysOverdue(null, '2026-06-01')).toBe(0);
  });
});

describe('getBillUrgency', () => {
  it('returns low when not overdue', () => {
    expect(getBillUrgency('2026-12-31', '2026-06-01')).toBe('low');
  });
  it('returns medium for 1-7 days', () => {
    expect(getBillUrgency('2026-05-28', '2026-06-01')).toBe('medium');
  });
  it('returns high for 8-30 days', () => {
    expect(getBillUrgency('2026-05-05', '2026-06-01')).toBe('high');
  });
  it('returns critical for over 30 days', () => {
    expect(getBillUrgency('2026-01-01', '2026-06-01')).toBe('critical');
  });
});
