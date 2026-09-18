import { describe, expect, it } from 'vitest';
import {
  createdTerminalId,
  filterTerminalsForLocation,
  mapBranchLocations,
} from './locationSelection';

describe('locationSelection', () => {
  it('maps only branches with valid positive ids', () => {
    expect(
      mapBranchLocations([
        { id: '3', ownerId: '1', name: 'Main', address: 'A', phone: '1', isActive: true, createdAt: '', updatedAt: '' },
        { id: '0', ownerId: '1', name: 'Broken', address: '', phone: '', isActive: true, createdAt: '', updatedAt: '' },
        { id: 'NaN', ownerId: '1', name: 'Missing', address: '', phone: '', isActive: true, createdAt: '', updatedAt: '' },
      ])
    ).toEqual([{ locationId: 3, name: 'Main', address: 'A', phone: '1' }]);
  });

  it('filters terminals by valid location id and active terminal id', () => {
    expect(
      filterTerminalsForLocation(
        [
          { id: '4', businessId: '1', branchId: '3', name: 'Register 1', deviceId: undefined, isActive: true, lastSeenAt: undefined, createdAt: '' },
          { id: '0', businessId: '1', branchId: '3', name: 'Broken', deviceId: undefined, isActive: true, lastSeenAt: undefined, createdAt: '' },
          { id: '5', businessId: '1', branchId: '9', name: 'Other', deviceId: undefined, isActive: true, lastSeenAt: undefined, createdAt: '' },
        ],
        3
      )
    ).toEqual([{ id: 4, name: 'Register 1', type: 'pos', isActive: true }]);
  });

  it('reads created terminal ids from generated responses', () => {
    expect(createdTerminalId('8')).toBe(8);
  });

  it('rejects created terminal responses without valid ids', () => {
    expect(() => createdTerminalId('draft')).toThrow(
      'Terminal registration did not return a valid terminal ID.',
    );
    expect(() => createdTerminalId('')).toThrow(
      'Terminal registration did not return a valid terminal ID.',
    );
  });
});
