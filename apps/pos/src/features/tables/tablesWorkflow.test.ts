import { describe, expect, it } from 'vitest';
import { mapRemoteTables } from './tablesWorkflow';

describe('mapRemoteTables', () => {
  it('maps remote tables with normalized fields', () => {
    expect(
      mapRemoteTables([
        {
          id: '7',
          label: ' T7 ',
          seats: 4,
          status: 'occupied',
        },
      ])
    ).toEqual([
      {
        id: 7,
        number: 'T7',
        seats: 4,
        status: 'seated',
      },
    ]);
  });

  it('drops tables without usable ids', () => {
    expect(mapRemoteTables([{ id: '0' }, { id: 'bad' }, {}])).toEqual([]);
  });

  it('defaults invalid status and seats', () => {
    expect(mapRemoteTables([{ id: '1', seats: -2, status: 'broken' }])).toEqual([
      {
        id: 1,
        number: '1',
        seats: 4,
        status: 'available',
      },
    ]);
  });

  it('maps dedicated table API labels and statuses', () => {
    expect(
      mapRemoteTables([
        { id: '2', label: ' Window ', seats: 3, status: 'reserved' },
        { id: '3', label: ' Patio ', seats: 5, status: 'cleaning' },
      ]),
    ).toEqual([
      {
        id: 2,
        number: 'Window',
        seats: 3,
        status: 'ordered',
      },
      {
        id: 3,
        number: 'Patio',
        seats: 5,
        status: 'dirty',
      },
    ]);
  });
});
