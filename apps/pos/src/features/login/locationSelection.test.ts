import { describe, expect, it } from 'vitest';
import {
  extractCreatedTerminalId,
  filterTerminalsForLocation,
  mapBranchLocations,
} from './locationSelection';

describe('locationSelection', () => {
  it('maps only branches with valid positive ids', () => {
    expect(
      mapBranchLocations([
        { id: 3, name: 'Main', address: 'A', phone: '1' },
        { id: 0, name: 'Broken' },
        { id: Number.NaN, name: 'Missing' },
      ])
    ).toEqual([{ locationId: 3, name: 'Main', address: 'A', phone: '1' }]);
  });

  it('filters terminals by valid location id and active terminal id', () => {
    expect(
      filterTerminalsForLocation(
        [
          { id: 4, name: 'Register 1', type: 'pos', branchId: 3 },
          { id: 0, name: 'Broken', type: 'pos', branchId: 3 },
          { id: 5, name: 'Other', type: 'pos', branchId: 9 },
        ],
        3
      )
    ).toEqual([{ id: 4, name: 'Register 1', type: 'pos', isActive: true }]);
  });

  it('extracts created terminal ids from known API response shapes', () => {
    expect(extractCreatedTerminalId({ id: '8' })).toBe(8);
    expect(extractCreatedTerminalId({ terminalId: 9 })).toBe(9);
    expect(extractCreatedTerminalId({ terminal: { id: '10' } })).toBe(10);
    expect(extractCreatedTerminalId({ data: { id: 11 } })).toBe(11);
  });

  it('rejects created terminal responses without valid ids', () => {
    expect(() => extractCreatedTerminalId({ id: 'draft' })).toThrow(
      'Terminal registration did not return a valid terminal ID.',
    );
    expect(() => extractCreatedTerminalId(null)).toThrow(
      'Terminal registration did not return a valid terminal ID.',
    );
  });
});
