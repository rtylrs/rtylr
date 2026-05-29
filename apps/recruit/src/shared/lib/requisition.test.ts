import { describe, it, expect } from 'vitest';
import {
  requisitionStatusLabel,
  requisitionPriorityLabel,
  openRequisitions,
  requisitionsByDepartment,
  urgentOpenRequisitions,
  timeToFill,
} from './requisition';
import type { JobRequisition } from './requisition';

function makeReq(
  status: 'draft' | 'open' | 'on_hold' | 'filled' | 'cancelled',
  priority: 'low' | 'medium' | 'high' | 'urgent' = 'medium',
  dept = 'Engineering',
  filledAt?: string
): JobRequisition {
  return {
    id: 'r1',
    title: 'Engineer',
    department: dept,
    status,
    priority,
    openedAt: '2026-01-01',
    filledAt,
  };
}

describe('requisitionStatusLabel', () => {
  it('returns display label for known statuses', () => {
    expect(requisitionStatusLabel('open')).toBe('Open');
    expect(requisitionStatusLabel('on_hold')).toBe('On Hold');
  });

  it('returns raw value for unknown status', () => {
    expect(requisitionStatusLabel('archived')).toBe('archived');
  });
});

describe('requisitionPriorityLabel', () => {
  it('returns display label for known priorities', () => {
    expect(requisitionPriorityLabel('urgent')).toBe('Urgent');
    expect(requisitionPriorityLabel('low')).toBe('Low');
  });
});

describe('openRequisitions', () => {
  it('filters only open requisitions', () => {
    const reqs = [makeReq('open'), makeReq('filled'), makeReq('draft')];
    expect(openRequisitions(reqs)).toHaveLength(1);
  });
});

describe('requisitionsByDepartment', () => {
  it('groups by department', () => {
    const reqs = [makeReq('open', 'medium', 'Engineering'), makeReq('open', 'medium', 'Sales')];
    const grouped = requisitionsByDepartment(reqs);
    expect(grouped['Engineering']).toHaveLength(1);
    expect(grouped['Sales']).toHaveLength(1);
  });
});

describe('urgentOpenRequisitions', () => {
  it('returns open urgent requisitions only', () => {
    const reqs = [
      makeReq('open', 'urgent'),
      makeReq('open', 'medium'),
      makeReq('filled', 'urgent'),
    ];
    expect(urgentOpenRequisitions(reqs)).toHaveLength(1);
  });
});

describe('timeToFill', () => {
  it('returns days from open to fill', () => {
    expect(timeToFill(makeReq('filled', 'medium', 'Engineering', '2026-02-10'))).toBe(40);
  });

  it('returns null when not filled', () => {
    expect(timeToFill(makeReq('open'))).toBeNull();
  });
});
