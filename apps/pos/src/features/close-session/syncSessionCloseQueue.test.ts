import { describe, expect, it, vi } from 'vitest';
import { syncQueuedSessionCloses } from './syncSessionCloseQueue';
import type { QueuedSessionClose } from '../../shared/stores/sessionCloseQueueStore';

const makeEntry = (overrides: Partial<QueuedSessionClose> = {}): QueuedSessionClose => ({
  id: 'sc_1',
  payload: {
    sessionId: 42,
    closingFloat: 95,
    expected: 100,
    variance: -5,
  },
  queuedAt: '2026-05-06T00:00:00.000Z',
  retries: 0,
  status: 'pending',
  ...overrides,
});

describe('syncQueuedSessionCloses', () => {
  it('marks pending session closes as synced after successful upload', async () => {
    const syncClose = vi.fn(async () => undefined);
    const retry = vi.fn();
    const markSynced = vi.fn();
    const markFailed = vi.fn();

    const result = await syncQueuedSessionCloses({
      entries: [makeEntry()],
      syncClose,
      retry,
      markSynced,
      markFailed,
    });

    expect(result).toEqual({ attempted: 1, synced: 1, failed: 0 });
    expect(retry).toHaveBeenCalledWith('sc_1');
    expect(syncClose).toHaveBeenCalledWith({
      sessionId: 42,
      closingFloat: 95,
      expected: 100,
      variance: -5,
    });
    expect(markSynced).toHaveBeenCalledWith('sc_1');
    expect(markFailed).not.toHaveBeenCalled();
  });

  it('marks failed uploads with the error message', async () => {
    const syncClose = vi.fn(async () => {
      throw new Error('offline');
    });
    const markFailed = vi.fn();

    const result = await syncQueuedSessionCloses({
      entries: [makeEntry()],
      syncClose,
      retry: vi.fn(),
      markSynced: vi.fn(),
      markFailed,
    });

    expect(result).toEqual({ attempted: 1, synced: 0, failed: 1 });
    expect(markFailed).toHaveBeenCalledWith('sc_1', 'offline');
  });
});
