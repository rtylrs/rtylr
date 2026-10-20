import { describe, expect, it, vi } from 'vitest';
import { detectDelimiter, parseCsv } from '../../../shared/lib/csv';
import {
  CHUNK_SIZE,
  guessMapping,
  type ImportTransport,
  mappingNamesLead,
  rowToLead,
  runLeadImport,
} from './leadImport';

describe('guessMapping', () => {
  it('maps common English headers', () => {
    const mapping = guessMapping([
      'First Name',
      'Last Name',
      'Company',
      'E-mail',
      'Phone Number',
      'Job Title',
      'Deal value',
    ]);
    expect(mapping).toEqual({
      firstName: 0,
      lastName: 1,
      companyName: 2,
      email: 3,
      phone: 4,
      jobTitle: 5,
      value: 6,
    });
  });

  it('maps French and Arabic headers, accents and hamza included', () => {
    expect(
      guessMapping(['Prénom', 'Nom de famille', 'Société', 'Téléphone', 'Courriel'])
    ).toMatchObject({
      firstName: 0,
      lastName: 1,
      companyName: 2,
      phone: 3,
      email: 4,
    });
    expect(
      guessMapping(['الاسم الأول', 'اسم العائلة', 'الشركة', 'رقم الهاتف', 'البريد الإلكتروني'])
    ).toMatchObject({
      firstName: 0,
      lastName: 1,
      companyName: 2,
      phone: 3,
      email: 4,
    });
  });

  it('uses a single name column only when there is no first name', () => {
    expect(guessMapping(['Name', 'Phone'])).toEqual({ fullName: 0, phone: 1 });
    expect(mappingNamesLead({ phone: 1 })).toBe(false);
    expect(mappingNamesLead({ fullName: 0 })).toBe(true);
  });
});

describe('rowToLead', () => {
  it('splits a full name and reads amounts with thousands separators', () => {
    const lead = rowToLead(
      ['Jane van Doe', '$1,500.5', 'Acme'],
      { fullName: 0, value: 1, companyName: 2 },
      'monthly'
    );
    expect(lead).toMatchObject({
      firstName: 'Jane',
      lastName: 'van Doe',
      value: '1500.50',
      billing: 'monthly',
      companyName: 'Acme',
    });
  });

  it('drops an unreadable value instead of failing the row', () => {
    expect(
      rowToLead(['Acme', 'call me'], { companyName: 0, value: 1 }, 'one_time').value
    ).toBeUndefined();
  });
});

describe('parseCsv', () => {
  it('detects semicolon and tab delimiters and strips a BOM', () => {
    expect(detectDelimiter('a;b;"c;d"\n1;2;3')).toBe(';');
    expect(parseCsv('﻿name;phone\n"Doe; Jane";123')).toEqual([
      ['name', 'phone'],
      ['Doe; Jane', '123'],
    ]);
    expect(parseCsv('name\tphone\nAcme\t555')).toEqual([
      ['name', 'phone'],
      ['Acme', '555'],
    ]);
  });
});

function transport(
  overrides: Partial<ImportTransport> = {}
): ImportTransport & { sent: number[][] } {
  const sent: number[][] = [];
  return {
    sent,
    start: vi.fn(async () => ({ id: '9' })),
    send: vi.fn(async (_id, body) => {
      sent.push((body.rows ?? []).map((row) => row.row ?? 0));
      return {
        results: (body.rows ?? []).map((row) =>
          row.lead?.companyName === 'Bad'
            ? { row: row.row, outcome: 'failed', error: 'no name' }
            : { row: row.row, outcome: 'created' }
        ),
      };
    }),
    finish: vi.fn(async () => ({})),
    ...overrides,
  };
}

describe('runLeadImport', () => {
  const options = {
    fileName: 'list.csv',
    duplicates: 'skip' as const,
    source: 'list',
    billing: 'one_time',
  };

  it('sends file rows in chunks with their line numbers and collects failures', async () => {
    const rows = Array.from({ length: CHUNK_SIZE + 3 }, (_, index) => [
      index === 1 ? 'Bad' : `Co ${index}`,
    ]);
    const fake = transport();
    const progress = vi.fn();
    const result = await runLeadImport(rows, { companyName: 0 }, options, fake, progress);

    expect(fake.sent).toHaveLength(2);
    expect(fake.sent[0][0]).toBe(2);
    expect(fake.sent[1]).toEqual([CHUNK_SIZE + 2, CHUNK_SIZE + 3, CHUNK_SIZE + 4]);
    expect(result.progress).toMatchObject({
      sent: CHUNK_SIZE + 3,
      created: CHUNK_SIZE + 2,
      failed: 1,
    });
    expect(result.failures).toEqual([{ line: 3, error: 'no name' }]);
    expect(fake.finish).toHaveBeenCalledWith('9');
    expect(progress).toHaveBeenLastCalledWith(expect.objectContaining({ sent: CHUNK_SIZE + 3 }));
  });

  it('retries a transient failure but not a rejected request', async () => {
    vi.useFakeTimers();
    try {
      let calls = 0;
      const flaky = transport({
        send: vi.fn(async (_id, body) => {
          calls += 1;
          if (calls === 1) throw Object.assign(new Error('bad gateway'), { statusCode: 502 });
          return {
            results: (body.rows ?? []).map((row) => ({ row: row.row, outcome: 'created' })),
          };
        }),
      });
      const run = runLeadImport([['Acme']], { companyName: 0 }, options, flaky, () => undefined);
      await vi.runAllTimersAsync();
      await expect(run).resolves.toMatchObject({ progress: { created: 1 } });
      expect(calls).toBe(2);

      const rejected = transport({
        send: vi.fn(async () => {
          throw Object.assign(new Error('closed'), { statusCode: 409 });
        }),
      });
      await expect(
        runLeadImport([['Acme']], { companyName: 0 }, options, rejected, () => undefined)
      ).rejects.toThrow('closed');
      expect(rejected.send).toHaveBeenCalledTimes(1);
    } finally {
      vi.useRealTimers();
    }
  });

  it('stops between chunks when cancelled and still finishes the import', async () => {
    const controller = new AbortController();
    const fake = transport();
    const rows = Array.from({ length: CHUNK_SIZE * 3 }, () => ['Acme']);
    const result = await runLeadImport(
      rows,
      { companyName: 0 },
      options,
      fake,
      (progress) => {
        if (progress.sent >= CHUNK_SIZE) controller.abort();
      },
      controller.signal
    );
    expect(fake.sent).toHaveLength(1);
    expect(result.progress.sent).toBe(CHUNK_SIZE);
    expect(fake.finish).toHaveBeenCalled();
  });
});
