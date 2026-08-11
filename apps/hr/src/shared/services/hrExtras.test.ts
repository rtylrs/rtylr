import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
  },
}));

import { hrExtras } from './hrExtras';
import { safe } from '../lib/api';

describe('hrExtras schedule calls', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // The API binds request fields by their snake_case JSON tag, so camelCase keys
  // arrive as zero values and the write silently does nothing.
  it('creates a schedule shift with snake_case keys', () => {
    hrExtras.upsertScheduleShift({
      staffId: 4101,
      staffName: 'Alice',
      role: 'Server',
      day: '2026-06-01',
      start: '09:00',
      end: '17:00',
    });
    expect(safe.post).toHaveBeenCalledWith('/hr/shifts', {
      staff_id: 4101,
      date: '2026-06-01',
      start_time: '09:00',
      end_time: '17:00',
    });
  });

  it('updates a schedule shift with snake_case keys', () => {
    hrExtras.upsertScheduleShift({
      id: 7,
      staffId: 4101,
      staffName: 'Alice',
      role: 'Server',
      day: '2026-06-02',
      start: '10:00',
      end: '18:00',
    });
    expect(safe.put).toHaveBeenCalledWith('/hr/shifts', {
      id: 7,
      date: '2026-06-02',
      start_time: '10:00',
      end_time: '18:00',
    });
  });

  // "published" is not a shift status. The column accepts only scheduled,
  // in_progress, completed, absent and cancelled; publishing belongs to a roster
  // assignment, so it must not be sent as a status on the shift row.
  it('never sends a status of published when upserting a shift', () => {
    hrExtras.upsertScheduleShift({
      staffId: 4101,
      staffName: 'Alice',
      role: 'Server',
      day: '2026-06-01',
      start: '09:00',
      end: '17:00',
      published: true,
    });
    const payload = (safe.post as unknown as ReturnType<typeof vi.fn>).mock.calls[0][1];
    expect(payload).not.toHaveProperty('status');
  });

  it('lists the schedule using the date filter the API binds', () => {
    hrExtras.listSchedule('2026-06-01');
    expect(safe.get).toHaveBeenCalledWith('/hr/shifts', { params: { date: '2026-06-01' } });
  });

  // Publishing is a POST to its own endpoint. This used to GET /hr/shifts with
  // view=publish, which no handler read, so it published nothing and still
  // reported success.
  it('publishes a week through the roster publish endpoint', () => {
    hrExtras.publishWeek('2026-09-07');
    expect(safe.post).toHaveBeenCalledWith('/hr/shifts/roster/publish', {
      week_start: '2026-09-07',
    });
  });

  it('deletes a schedule shift by id', () => {
    hrExtras.deleteScheduleShift(19);
    expect(safe.delete).toHaveBeenCalledWith('/hr/shifts', { data: { id: 19 } });
  });
});
