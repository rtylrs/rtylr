import { describe, expect, it } from 'vitest';
import { mapRemoteTables, normalizeTablesResponse } from './tablesWorkflow';

describe('mapRemoteTables', () => {
  it('maps remote tables with normalized fields', () => {
    expect(
      mapRemoteTables([
        {
          id: 7.9,
          number: ' T7 ',
          seats: '4.8',
          status: 'occupied',
          guests: Number.NaN,
          total: -5,
          server: ' Alice ',
        },
      ])
    ).toEqual([
      {
        id: 7,
        number: 'T7',
        seats: 4,
        status: 'seated',
        guests: undefined,
        total: 0,
        server: 'Alice',
      },
    ]);
  });

  it('drops tables without usable ids', () => {
    expect(mapRemoteTables([null, { id: 0 }, { id: Number.NaN }, { number: 'T1' }])).toEqual([]);
  });

  it('defaults invalid status and seats', () => {
    expect(mapRemoteTables([{ id: 1, seats: -2, status: 'broken' }])).toEqual([
      {
        id: 1,
        number: '1',
        seats: 4,
        status: 'available',
        guests: undefined,
        total: undefined,
        server: undefined,
      },
    ]);
  });

  it('maps dedicated table API labels and statuses', () => {
    expect(
      mapRemoteTables([
        { id: 2, label: ' Window ', seats: 3, status: 'reserved' },
        { id: 3, label: ' Patio ', seats: 5, status: 'cleaning' },
      ]),
    ).toEqual([
      {
        id: 2,
        number: 'Window',
        seats: 3,
        status: 'ordered',
        guests: undefined,
        total: undefined,
        server: undefined,
      },
      {
        id: 3,
        number: 'Patio',
        seats: 5,
        status: 'dirty',
        guests: undefined,
        total: undefined,
        server: undefined,
      },
    ]);
  });

  it('normalizes common table API response envelopes', () => {
    expect(
      normalizeTablesResponse({
        data: {
          tables: [{ id: '9', name: ' Patio ', seats: '2', status: 'dirty' }],
        },
      }),
    ).toEqual([
      {
        id: 9,
        number: 'Patio',
        seats: 2,
        status: 'dirty',
        guests: undefined,
        total: undefined,
        server: undefined,
      },
    ]);
  });

  it('returns an empty list for non-array table responses', () => {
    expect(normalizeTablesResponse({ data: { tables: 'bad' } })).toEqual([]);
    expect(normalizeTablesResponse(null)).toEqual([]);
  });
});
