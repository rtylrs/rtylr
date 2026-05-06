import { describe, expect, it } from 'vitest';
import { buildCartContext } from './cartContext';

describe('buildCartContext', () => {
  it('builds numeric cart context from valid ids', () => {
    expect(buildCartContext(3.9, 5.2, '7')).toEqual({
      branchId: 3,
      terminalId: 5,
      employeeId: 7,
    });
  });

  it('returns null when location or terminal is missing', () => {
    expect(buildCartContext(null, 5, '7')).toBeNull();
    expect(buildCartContext(3, null, '7')).toBeNull();
  });

  it('returns null when employee id is not numeric', () => {
    expect(buildCartContext(3, 5, 'employee-uuid')).toBeNull();
  });
});
