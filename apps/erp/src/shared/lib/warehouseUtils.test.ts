import { describe, it, expect } from 'vitest';
import {
  warehouseUsedPct,
  warehouseFreeCapacity,
  isWarehouseNearCapacity,
  isWarehouseFull,
  sortWarehousesDefaultFirst,
  findDefaultWarehouse,
  warehouseCapacityLabel,
} from './warehouseUtils';
import type { Warehouse } from '../services/erpApi';

const base: Warehouse = {
  id: 1,
  name: 'Main',
  address: '123 St',
  manager: 'Alice',
  capacity: 1000,
  used: 750,
  isDefault: true,
};

const secondary: Warehouse = {
  id: 2,
  name: 'Secondary',
  address: '456 Ave',
  manager: 'Bob',
  capacity: 500,
  used: 500,
  isDefault: false,
};

describe('warehouseUsedPct', () => {
  it('returns correct percentage', () => {
    expect(warehouseUsedPct(base)).toBe(75);
  });

  it('returns 0 when capacity is 0', () => {
    expect(warehouseUsedPct({ ...base, capacity: 0 })).toBe(0);
  });

  it('clamps to 100 when over capacity', () => {
    expect(warehouseUsedPct({ ...base, used: 1200, capacity: 1000 })).toBe(100);
  });

  it('returns 100 when at full capacity', () => {
    expect(warehouseUsedPct(secondary)).toBe(100);
  });
});

describe('warehouseFreeCapacity', () => {
  it('returns remaining capacity', () => {
    expect(warehouseFreeCapacity(base)).toBe(250);
  });

  it('returns 0 when at full capacity', () => {
    expect(warehouseFreeCapacity(secondary)).toBe(0);
  });

  it('returns 0 when used exceeds capacity', () => {
    expect(warehouseFreeCapacity({ ...base, used: 1200, capacity: 1000 })).toBe(0);
  });
});

describe('isWarehouseNearCapacity', () => {
  it('returns true when at or above 90% threshold', () => {
    expect(isWarehouseNearCapacity({ ...base, used: 900, capacity: 1000 })).toBe(true);
    expect(isWarehouseNearCapacity(secondary)).toBe(true);
  });

  it('returns false when below 90% threshold', () => {
    expect(isWarehouseNearCapacity(base)).toBe(false);
  });

  it('accepts custom threshold', () => {
    expect(isWarehouseNearCapacity(base, 75)).toBe(true);
    expect(isWarehouseNearCapacity(base, 80)).toBe(false);
  });
});

describe('isWarehouseFull', () => {
  it('returns true when used equals capacity', () => {
    expect(isWarehouseFull(secondary)).toBe(true);
  });

  it('returns false when not full', () => {
    expect(isWarehouseFull(base)).toBe(false);
  });
});

describe('sortWarehousesDefaultFirst', () => {
  it('puts default warehouse first', () => {
    const sorted = sortWarehousesDefaultFirst([secondary, base]);
    expect(sorted[0].isDefault).toBe(true);
  });

  it('sorts non-default warehouses alphabetically', () => {
    const alpha: Warehouse = { ...secondary, id: 3, name: 'Alpha' };
    const sorted = sortWarehousesDefaultFirst([secondary, alpha]);
    expect(sorted[0].name).toBe('Alpha');
  });

  it('does not mutate original array', () => {
    const original = [secondary, base];
    sortWarehousesDefaultFirst(original);
    expect(original[0].name).toBe('Secondary');
  });
});

describe('findDefaultWarehouse', () => {
  it('returns default warehouse', () => {
    expect(findDefaultWarehouse([secondary, base])?.isDefault).toBe(true);
  });

  it('returns undefined when no default', () => {
    expect(findDefaultWarehouse([secondary])).toBeUndefined();
  });
});

describe('warehouseCapacityLabel', () => {
  it('returns formatted capacity label', () => {
    expect(warehouseCapacityLabel(base)).toBe('750 / 1,000');
  });

  it('handles zero used', () => {
    expect(warehouseCapacityLabel({ ...base, used: 0 })).toBe('0 / 1,000');
  });
});
