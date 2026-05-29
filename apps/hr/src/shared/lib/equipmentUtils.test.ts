import { describe, it, expect } from 'vitest';
import {
  equipmentStatusLabel,
  equipmentStatusColor,
  isAssignable,
  isAssigned,
  isActiveEquipment,
  depreciatedValue,
  countAssigned,
  countAvailable,
  filterByEquipmentCategory,
  utilizationRate,
} from './equipmentUtils';

describe('equipmentStatusLabel', () => {
  it('returns Available for Available', () => {
    expect(equipmentStatusLabel('Available')).toBe('Available');
  });
  it('returns In Repair for In Repair', () => {
    expect(equipmentStatusLabel('In Repair')).toBe('In Repair');
  });
  it('returns Retired for Retired', () => {
    expect(equipmentStatusLabel('Retired')).toBe('Retired');
  });
});

describe('equipmentStatusColor', () => {
  it('returns green for Available', () => {
    expect(equipmentStatusColor('Available')).toBe('green');
  });
  it('returns blue for Assigned', () => {
    expect(equipmentStatusColor('Assigned')).toBe('blue');
  });
  it('returns gray for Retired', () => {
    expect(equipmentStatusColor('Retired')).toBe('gray');
  });
});

describe('isAssignable', () => {
  it('returns true for Available', () => {
    expect(isAssignable('Available')).toBe(true);
  });
  it('returns false for Assigned', () => {
    expect(isAssignable('Assigned')).toBe(false);
  });
  it('returns false for Retired', () => {
    expect(isAssignable('Retired')).toBe(false);
  });
});

describe('isAssigned', () => {
  it('returns true for Assigned', () => {
    expect(isAssigned('Assigned')).toBe(true);
  });
  it('returns false for Available', () => {
    expect(isAssigned('Available')).toBe(false);
  });
});

describe('isActiveEquipment', () => {
  it('returns true for Available', () => {
    expect(isActiveEquipment('Available')).toBe(true);
  });
  it('returns true for Assigned', () => {
    expect(isActiveEquipment('Assigned')).toBe(true);
  });
  it('returns false for Retired', () => {
    expect(isActiveEquipment('Retired')).toBe(false);
  });
});

describe('depreciatedValue', () => {
  it('calculates depreciated value correctly', () => {
    expect(depreciatedValue(1000, 2, 5)).toBe(600);
  });
  it('returns 0 when fully depreciated', () => {
    expect(depreciatedValue(1000, 5, 5)).toBe(0);
  });
  it('returns 0 when over useful life', () => {
    expect(depreciatedValue(1000, 10, 5)).toBe(0);
  });
  it('returns full value when age is 0', () => {
    expect(depreciatedValue(1000, 0, 5)).toBe(1000);
  });
});

describe('countAssigned', () => {
  it('counts assigned records', () => {
    const records = [
      { status: 'Assigned' as const },
      { status: 'Available' as const },
      { status: 'Assigned' as const },
    ];
    expect(countAssigned(records)).toBe(2);
  });
});

describe('countAvailable', () => {
  it('counts available records', () => {
    const records = [
      { status: 'Available' as const },
      { status: 'Retired' as const },
      { status: 'Available' as const },
    ];
    expect(countAvailable(records)).toBe(2);
  });
});

describe('filterByEquipmentCategory', () => {
  it('filters by category', () => {
    const records = [
      { category: 'Laptop' as const, id: 1 },
      { category: 'Phone' as const, id: 2 },
      { category: 'Laptop' as const, id: 3 },
    ];
    const result = filterByEquipmentCategory(records, 'Laptop');
    expect(result).toHaveLength(2);
  });
});

describe('utilizationRate', () => {
  it('calculates utilization correctly', () => {
    const records = [{ status: 'Assigned' as const }, { status: 'Available' as const }];
    expect(utilizationRate(records)).toBe(0.5);
  });
  it('returns 0 when no active equipment', () => {
    const records = [{ status: 'Retired' as const }];
    expect(utilizationRate(records)).toBe(0);
  });
  it('returns 1 when all assigned', () => {
    const records = [{ status: 'Assigned' as const }, { status: 'Assigned' as const }];
    expect(utilizationRate(records)).toBe(1);
  });
});
